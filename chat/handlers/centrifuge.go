package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/centrifugal/centrifuge"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"nkonev.name/chat/auth"
	"nkonev.name/chat/db"
	"nkonev.name/chat/handlers/dto"
	. "nkonev.name/chat/logger"
	"nkonev.name/chat/utils"
	"strings"
	"time"
)

func handleLog(e centrifuge.LogEntry) {
	Logger.Printf("%s: %v", e.Message, e.Fields)
}

func getChanPresenceStats(presenceManager centrifuge.PresenceManager, client *centrifuge.Client, e interface{}) *centrifuge.PresenceStats {
	var channel string
	switch v := e.(type) {
	case centrifuge.SubscribeEvent:
		channel = v.Channel
		break
	case centrifuge.UnsubscribeEvent:
		channel = v.Channel
		break
	default:
		Logger.Errorf("Unknown type of event")
		return nil
	}
	stats, err := presenceManager.PresenceStats(channel)
	if err != nil {
		Logger.Errorf("Error during get stats %v", err)
	}
	Logger.Printf("client id=%v, userId=%v acting with channel %s, channelStats.NumUsers %v", client.ID(), client.UserID(), channel, stats.NumUsers)
	return &stats
}

type PassData struct {
	Payload  utils.H `json:"payload"`
	Metadata utils.H `json:"metadata"`
}

type TypedMessage struct {
	Type string `json:"type"`
}

type MessageRead struct {
	ChatId    int64 `json:"chatId"`
	MessageId int64 `json:"messageId"`
}

// clientId - temporary (session?) UUID, generated by centrifuge
// userId - permanent user id stored in database
func modifyMessage(msg []byte, originatorUserId string, originatorClientId string) ([]byte, error) {
	var v = &PassData{}
	if err := json.Unmarshal(msg, v); err != nil {
		return nil, err
	}
	v.Metadata = utils.H{"originatorUserId": originatorUserId, "originatorClientId": originatorClientId}
	return json.Marshal(v)
}

func getNewMessagesNotificationMessage(dbs db.DB, userId int64) ([]byte, error) {
	count, err := dbs.GetAllUnreadMessagesCount(userId)
	if err != nil {
		Logger.Errorf("Unable to get unread messages count from db %v", err)
		return nil, err
	} else {
		var mc = dto.AllUnreadMessages{MessagesCount: count}
		if marshal, err := json.Marshal(&mc); err != nil {
			Logger.Errorf("Unable to marshall unread messages count %v", err)
			return nil, err
		} else {
			return marshal, nil
		}
	}
}

func ConfigureCentrifuge(lc fx.Lifecycle, dbs db.DB) *centrifuge.Node {
	// We use default config here as starting point. Default config contains
	// reasonable values for available options.
	cfg := centrifuge.DefaultConfig

	// Centrifuge library exposes logs with different log level. In your app
	// you can set special function to handle these log entries in a way you want.
	cfg.LogLevel = centrifuge.LogLevelDebug
	cfg.LogHandler = handleLog

	// Node is the core object in Centrifuge library responsible for many useful
	// things. Here we initialize new Node instance and pass config to it.
	node, _ := centrifuge.New(cfg)

	redisAddress := viper.GetString("centrifuge.redis.address")
	redisPassword := viper.GetString("centrifuge.redis.password")
	redisDB := viper.GetInt("centrifuge.redis.db")
	readTimeout := viper.GetDuration("centrifuge.redis.readTimeout")
	writeTimeout := viper.GetDuration("centrifuge.redis.writeTimeout")
	connectTimeout := viper.GetDuration("centrifuge.redis.connectTimeout")
	idleTimeout := viper.GetDuration("centrifuge.redis.idleTimeout")

	redisShardConfigs := []centrifuge.RedisShardConfig{
		{Address: redisAddress, DB: redisDB, Password: redisPassword, IdleTimeout: idleTimeout, ReadTimeout: readTimeout, ConnectTimeout: connectTimeout, WriteTimeout: writeTimeout},
	}
	var redisShards []*centrifuge.RedisShard
	for _, redisConf := range redisShardConfigs {
		redisShard, err := centrifuge.NewRedisShard(node, redisConf)
		if err != nil {
			Logger.Fatal(err)
		}
		redisShards = append(redisShards, redisShard)
	}

	broker, err := centrifuge.NewRedisBroker(node, centrifuge.RedisBrokerConfig{
		// Use reasonably large expiration interval for stream meta key,
		// much bigger than maximum HistoryLifetime value in Node config.
		// This way stream meta data will expire, in some cases you may want
		// to prevent its expiration setting this to zero value.
		HistoryMetaTTL: 24 * time.Hour,

		// And configure a couple of shards to use.
		Shards: redisShards,
		Prefix: "centrifuge",
	})
	if err != nil {
		Logger.Fatal(err)
	}
	node.SetBroker(broker)

	presenceManager, err := centrifuge.NewRedisPresenceManager(node, centrifuge.RedisPresenceManagerConfig{
		Shards: redisShards,
	})
	if err != nil {
		Logger.Fatal(err)
	}
	node.SetPresenceManager(presenceManager)

	node.OnConnect(func(client *centrifuge.Client) {
		// Set Subscribe Handler to react on every channel subscription attempt
		// initiated by client. Here you can theoretically return an error or
		// disconnect client from server if needed. But now we just accept
		// all subscriptions.

		var creds, ok = centrifuge.GetCredentials(client.Context())
		if !ok {
			Logger.Infof("Cannot extract credentials")
			return
		}
		Logger.Infof("Connected websocket centrifuge client hasCredentials %v, credentials.userId=%v, credentials.expireAt=%v", ok, creds.UserID, creds.ExpireAt)
		userId, err := utils.ParseInt64(creds.UserID)
		if err != nil {
			Logger.Errorf("Unable to parse userId from %v", creds.UserID)
			return
		}
		var authResult = auth.AuthResult{}
		err = json.Unmarshal(creds.Info, &authResult)
		if err != nil {
			Logger.Errorf("Unable to parse authResult from creds: %v", err)
			return
		}

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			chatId, channelName, err := getChatId(e.Channel)
			if err != nil {
				Logger.Errorf("Error getting channel id %v", err)
				cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
				return
			}
			Logger.Infof("Get channel id %v, channel name %v", chatId, channelName)

			err = checkPermissions(dbs, creds.UserID, chatId, channelName)
			if err != nil {
				Logger.Errorf("Error during checking permissions userId %v, channelId %v, channelName %v,", creds.UserID, chatId, channelName)
				cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
				return
			}

			Logger.Printf("user %s subscribes on %s", client.UserID(), e.Channel)
			cb(centrifuge.SubscribeReply{
				Options: centrifuge.SubscribeOptions{
					Presence:  true,
					JoinLeave: true,
					Recover:   true,
				},
			}, nil)
		})

		client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
			Logger.Infof("user %s unsubscribed from %s", client.UserID(), e.Channel)
			getChanPresenceStats(presenceManager, client, e)
			return
		})

		// Set Publish Handler to react on every channel Publication sent by client.
		// Inside this method you can validate client permissions to publish into
		// channel. But in our simple chat app we allow everyone to publish into
		// any channel.
		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			Logger.Printf("User %v publishes into channel %s: %s", creds.UserID, e.Channel, string(e.Data))
			message, err := modifyMessage(e.Data, e.ClientInfo.UserID, e.ClientInfo.ClientID)
			if err != nil {
				Logger.Errorf("Error during modifyMessage %v", err)
				cb(centrifuge.PublishReply{}, centrifuge.ErrorInternal)
				return
			}
			result, err := node.Publish(
				e.Channel, message,
				centrifuge.WithHistory(300, time.Minute),
				centrifuge.WithClientInfo(e.ClientInfo),
			)
			if err != nil {
				Logger.Errorf("Error during publishing modified message %v", err)
			}

			cb(centrifuge.PublishReply{Result: &result}, err)
		})

		client.OnPresence(func(e centrifuge.PresenceEvent, cb centrifuge.PresenceCallback) {
			cb(centrifuge.PresenceReply{}, nil)
		})

		// Set Disconnect Handler to react on client disconnect events.
		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			Logger.Printf("Centrifuge user %v disconnected", creds.UserID)
		})

		client.OnRefresh(func(event centrifuge.RefreshEvent, cb centrifuge.RefreshCallback) {
			expirationTimestamp := getNextRefreshTime()
			Logger.Infof("Refreshing centrifuge session for user %v", userId)
			cb(centrifuge.RefreshReply{
				ExpireAt: expirationTimestamp,
			}, nil)
		})

		client.OnSubRefresh(func(event centrifuge.SubRefreshEvent, cb centrifuge.SubRefreshCallback) {
			expirationTimestamp := getNextRefreshTime()
			Logger.Infof("SubRefreshing centrifuge subscription for user %v", userId)
			cb(centrifuge.SubRefreshReply{
				ExpireAt: expirationTimestamp,
			}, nil)
		})

		client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
			var reply = centrifuge.RPCReply{}
			if event.Method == "check_for_new_messages" {
				message, err := getNewMessagesNotificationMessage(dbs, userId)
				reply.Data = message
				callback(reply, err)
			} else if event.Method == "message_read" {
				mr := MessageRead{}
				if err = json.Unmarshal(event.Data, &mr); err != nil {
					Logger.Errorf("client %v sent non-parseable message - %v", creds.UserID, err)
					callback(reply, err)
				} else {
					// TODO to separated centrifuge messages handler
					Logger.Infof("Putting message read messageId=%v, chatId=%v, userId=%v", mr.MessageId, mr.ChatId, userId)
					err = markMessageAsRead(dbs, userId, mr.ChatId, mr.MessageId)
					if err != nil {
						Logger.Errorf("Error during putting message read messageId=%v, chatId=%v, userId=%v: err=%v", mr.MessageId, mr.ChatId, userId, err)
						callback(reply, err)
						return
					}
					message, err := getNewMessagesNotificationMessage(dbs, userId)
					reply.Data = message
					callback(reply, err)
				}
			} else {
				Logger.Errorf("Unknown method %v", event.Method)
				callback(reply, errors.New("Unknown method"))
			}
		})

		// In our example transport will always be Websocket but it can also be SockJS.
		transportName := client.Transport().Name()
		// In our example clients connect with JSON protocol but it can also be Protobuf.
		transportProtocol := client.Transport().Protocol()
		Logger.Printf("Centrifuge user %v connected via %s (%s)", creds.UserID, transportName, transportProtocol)
	})

	node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		cred, _ := centrifuge.GetCredentials(ctx)
		return centrifuge.ConnectReply{
			Data: []byte(`{}`),
			// Subscribe to personal several server-side channel.
			Subscriptions: map[string]centrifuge.SubscribeOptions{
				utils.PersonalChannelPrefix + cred.UserID: {Recover: true, Presence: true, JoinLeave: true},
			},
		}, nil
	})

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// do some work on application stop (like closing connections and files)
			Logger.Infof("Stopping centrifuge")
			return node.Shutdown(ctx)
		},
	})

	return node
}

func getNextRefreshTime() int64 {
	return time.Now().Unix() + 10*60
}

func checkPermissions(dbs db.DB, userId string, channelId int64, channelName string) error {
	if utils.CHANNEL_PREFIX_CHAT_MESSAGES == channelName {
		if ids, err := dbs.GetParticipantIds(channelId); err != nil {
			return err
		} else {
			for _, uid := range ids {
				if fmt.Sprintf("%v", uid) == userId {
					Logger.Infof("User %v found among participants of chat %v", userId, channelId)
					return nil
				}
			}
			return errors.New(fmt.Sprintf("User %v not found among participants", userId))
		}
	}
	return errors.New(fmt.Sprintf("User %v not allowed to use unknown channel %v", userId, channelName))
}

func getChatId(channel string) (int64, string, error) {
	if strings.HasPrefix(channel, utils.CHANNEL_PREFIX_CHAT_MESSAGES) {
		s := channel[len(utils.CHANNEL_PREFIX_CHAT_MESSAGES):]
		if parseInt64, err := utils.ParseInt64(s); err != nil {
			return 0, "", err
		} else {
			return parseInt64, utils.CHANNEL_PREFIX_CHAT_MESSAGES, nil
		}
	} else {
		return 0, "", errors.New("Subscription to unexpected channel: '" + channel + "'")
	}
}

func markMessageAsRead(db db.DB, userId, chatId, messageId int64) error {
	if participant, err := db.IsParticipant(userId, chatId); err != nil {
		Logger.Errorf("Error during checking participant")
		return err
	} else if !participant {
		Logger.Infof("User %v is not participant of chat %v, skipping", userId, chatId)
		return errors.New("Not authorized")
	}

	if err := db.AddMessageRead(messageId, userId, chatId); err != nil {
		return err
	}
	return nil
}
