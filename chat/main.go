package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/araddon/dateparse"
	"github.com/centrifugal/centrifuge"
	"github.com/centrifugal/protocol"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nkonev/videochat/client"
	. "github.com/nkonev/videochat/logger"
	"github.com/nkonev/videochat/utils"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"time"
)

type staticMiddleware echo.MiddlewareFunc
type authMiddleware echo.MiddlewareFunc

func main() {
	configFile := utils.InitFlags("./chat/config-dev/config.yml")
	utils.InitViper(configFile, "VIDEOCHAT")

	app := fx.New(
		fx.Logger(Logger),
		fx.Provide(
			client.NewRestClient,
			configureCentrifuge,
			configureEcho,
			configureStaticMiddleware,
			configureAuthMiddleware,
		),
		fx.Invoke(runCentrifuge, runEcho),
	)
	app.Run()

	Logger.Infof("Exit program")
}

func handleLog(e centrifuge.LogEntry) {
	Logger.Printf("%s: %v", e.Message, e.Fields)
}

func getChanPresenceStats(engine centrifuge.Engine, client *centrifuge.Client, e interface{}) *centrifuge.PresenceStats {
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
	stats, err := engine.PresenceStats(channel)
	if err != nil {
		Logger.Errorf("Error during get stats %v", err)
	}
	Logger.Printf("client id=%v, userId=%v subscribes on channel %s, channelStats.NumUsers %v", client.ID(), client.UserID(), channel, stats.NumUsers)
	return &stats
}

func configureCentrifuge(lc fx.Lifecycle) *centrifuge.Node {
	// We use default config here as starting point. Default config contains
	// reasonable values for available options.
	cfg := centrifuge.DefaultConfig
	// In this example we want client to do all possible actions with server
	// without any authentication and authorization. Insecure flag DISABLES
	// many security related checks in library. This is only to make example
	// short. In real app you most probably want authenticate and authorize
	// access to server. See godoc and examples in repo for more details.
	cfg.ClientInsecure = false
	// By default clients can not publish messages into channels. Setting this
	// option to true we allow them to publish.
	cfg.Publish = true

	// Centrifuge library exposes logs with different log level. In your app
	// you can set special function to handle these log entries in a way you want.
	cfg.LogLevel = centrifuge.LogLevelDebug
	cfg.LogHandler = handleLog

	// Node is the core object in Centrifuge library responsible for many useful
	// things. Here we initialize new Node instance and pass config to it.
	node, _ := centrifuge.New(cfg)

	engine, _ := centrifuge.NewMemoryEngine(node, centrifuge.MemoryEngineConfig{})
	node.SetEngine(engine)

	// ClientConnected node event handler is a point where you generally create a
	// binding between Centrifuge and your app business logic. Callback function you
	// pass here will be called every time new connection established with server.
	// Inside this callback function you can set various event handlers for connection.
	node.On().ClientConnected(func(ctx context.Context, client *centrifuge.Client) {
		// Set Subscribe Handler to react on every channel subscribtion attempt
		// initiated by client. Here you can theoretically return an error or
		// disconnect client from server if needed. But now we just accept
		// all subscriptions.

		client.On().Subscribe(func(e centrifuge.SubscribeEvent) centrifuge.SubscribeReply {
			// TODO make same duration as session
			presenceDuration, _ := time.ParseDuration("24h")
			clientInfo := &protocol.ClientInfo{
				User:   client.ID(),
				Client: client.UserID(),
			}
			err := engine.AddPresence(e.Channel, client.UserID(), clientInfo, presenceDuration)
			if err != nil {
				Logger.Errorf("Error during AddPresence %v", err)
			}

			if e.Channel == "aux" {
				stats := getChanPresenceStats(engine, client, e)

				type AuxChannelRequest struct {
					MessageType string `json:"type"`
				}

				if stats.NumUsers == 1 {
					data, _ := json.Marshal(AuxChannelRequest{"created"})
					Logger.Infof("Publishing created to channel %v", e.Channel)
					//err := node.Publish(e.Channel, data)
					err := client.Send(data)
					if err != nil {
						Logger.Errorf("Error during publishing created %v", err)
					}
				} else if stats.NumUsers > 1 {
					data, _ := json.Marshal(AuxChannelRequest{"joined"})
					Logger.Infof("Publishing joined to channel %v", e.Channel)
					// send to existing subscribers
					err := node.Publish(e.Channel, data)
					if err != nil {
						Logger.Errorf("Error during publishing joined %v", err)
					}
					// send to just subscribing client
					err2 := client.Send(data)
					if err2 != nil {
						Logger.Errorf("Error during publishing joined %v", err2)
					}

				}
			}

			return centrifuge.SubscribeReply{}
		})

		client.On().Unsubscribe(func(e centrifuge.UnsubscribeEvent) centrifuge.UnsubscribeReply {
			err := engine.RemovePresence(e.Channel, client.UserID())
			if err != nil {
				Logger.Errorf("Error during RemovePresence %v", err)
			}
			getChanPresenceStats(engine, client, e)

			return centrifuge.UnsubscribeReply{}
		})

		// Set Publish Handler to react on every channel Publication sent by client.
		// Inside this method you can validate client permissions to publish into
		// channel. But in our simple chat app we allow everyone to publish into
		// any channel.
		client.On().Publish(func(e centrifuge.PublishEvent) centrifuge.PublishReply {
			Logger.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
			return centrifuge.PublishReply{}
		})

		// Set Disconnect Handler to react on client disconnect events.
		client.On().Disconnect(func(e centrifuge.DisconnectEvent) centrifuge.DisconnectReply {
			Logger.Printf("client disconnected")
			return centrifuge.DisconnectReply{}
		})

		// In our example transport will always be Websocket but it can also be SockJS.
		transportName := client.Transport().Name()
		// In our example clients connect with JSON protocol but it can also be Protobuf.
		transportEncoding := client.Transport().Encoding()

		Logger.Printf("client connected via %s (%s)", transportName, transportEncoding)
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

func runCentrifuge(node *centrifuge.Node) {
	// Run node.
	Logger.Infof("Starting centrifuge...")
	go func() {
		if err := node.Run(); err != nil {
			Logger.Fatalf("Error on start centrifuge: %v", err)
		}
	}()
	Logger.Info("Centrifuge started.")
}

type authResult struct {
	UserId    string `json:"id"`
	UserLogin string `json:"login"`
	ExpiresAt int64  `json:"expiresAt"` // in GMT. in seconds
}

func extractAuth(request *http.Request) (*authResult, error) {
	expiresInString := request.Header.Get("X-Auth-Expiresin")
	t, err := dateparse.ParseLocal(expiresInString)
	if err != nil {
		return nil, err
	}
	return &authResult{
		UserId:    request.Header.Get("X-Auth-UserId"),
		UserLogin: request.Header.Get("X-Auth-Username"),
		ExpiresAt: t.Unix(),
	}, nil
}

// https://www.keycloak.org/docs/latest/securing_apps/index.html#upstream-headers
// authorize checks authentication of each requests (websocket establishment or regular ones)
//
// Parameters:
//
//  - `request` : http request to check
//  - `httpClient` : client to check authorization
//
// Returns:
//
//  - *authResult pointer or nil
//  - is whitelisted
//  - error
func authorize(request *http.Request) (*authResult, bool, error) {
	whitelistStr := viper.GetStringSlice("auth.exclude")
	whitelist := utils.StringsToRegexpArray(whitelistStr)
	if utils.CheckUrlInWhitelist(whitelist, request.RequestURI) {
		return nil, true, nil
	}
	auth, err := extractAuth(request)
	if err != nil {
		Logger.Infof("Error during extract authResult: %v", err)
		return nil, false, nil
	}
	GetLogEntry(request).Infof("Success authResult: %v", *auth)
	return auth, false, nil
}

func configureAuthMiddleware() authMiddleware {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authResult, whitelist, err := authorize(c.Request())
			if err != nil {
				Logger.Errorf("Error during authorize: %v", err)
				return err
			} else if whitelist {
				return next(c)
			} else if authResult == nil {
				return c.JSON(http.StatusUnauthorized, &utils.H{"status": "unauthorized"})
			} else {
				c.Set(utils.USER_PRINCIPAL_DTO, authResult)
				return next(c)
			}
		}
	}
}

func centrifugeAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authResult, _, err := authorize(r)
		if err != nil {
			Logger.Errorf("Error during try to authenticate centrifuge request: %v", err)
			return
		} else if authResult == nil {
			Logger.Errorf("Not authenticated centrifuge request")
			return
		} else {
			ctx := r.Context()
			newCtx := centrifuge.SetCredentials(ctx, &centrifuge.Credentials{
				UserID:   authResult.UserId,
				ExpireAt: authResult.ExpiresAt,
				Info:     []byte(fmt.Sprintf("{\"login\": \"%v\"}", authResult.UserLogin)),
			})
			r = r.WithContext(newCtx)
			h.ServeHTTP(w, r)
		}
	})
}

func configureEcho(staticMiddleware staticMiddleware, authMiddleware authMiddleware, lc fx.Lifecycle, node *centrifuge.Node) *echo.Echo {
	bodyLimit := viper.GetString("server.body.limit")

	e := echo.New()
	e.Logger.SetOutput(Logger.Writer())

	e.Pre(echo.MiddlewareFunc(staticMiddleware))
	e.Use(echo.MiddlewareFunc(authMiddleware))
	accessLoggerConfig := middleware.LoggerConfig{
		Output: Logger.Writer(),
		Format: `"remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out},"traceId":"${header:X-B3-Traceid}"` + "\n",
	}
	e.Use(middleware.LoggerWithConfig(accessLoggerConfig))
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit(bodyLimit))

	e.GET("/api/chat/websocket", convert(centrifugeAuthMiddleware(centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{}))))
	e.GET("/api/chat", chatHandler)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// do some work on application stop (like closing connections and files)
			Logger.Infof("Stopping server")
			return e.Shutdown(ctx)
		},
	})

	return e
}

func convert(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

type ChatDto struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func chatHandler(c echo.Context) error {
	usrs := []ChatDto{
		ChatDto{Name: "With Terry", Id: 1},
		ChatDto{Name: "Friday drunk", Id: 2},
		ChatDto{Name: "Lorem", Id: 3},
		ChatDto{Name: "Impsum", Id: 4},
		ChatDto{Name: "Dolor", Id: 5},
		ChatDto{Name: "With collegues", Id: 6},
		ChatDto{Name: "Lorem", Id: 7},
		ChatDto{Name: "Impsum", Id: 8},
		ChatDto{Name: "Dolor", Id: 9},
		ChatDto{Name: "With collegues", Id: 10},
		ChatDto{Name: "Lorem", Id: 11},
		ChatDto{Name: "Impsum", Id: 12},
		ChatDto{Name: "Dolor", Id: 13},
		ChatDto{Name: "With collegues", Id: 14},
		ChatDto{Name: "Lorem", Id: 15},
		ChatDto{Name: "Impsum", Id: 16},
		ChatDto{Name: "Dolor", Id: 17},
		ChatDto{Name: "With collegues", Id: 18},
		ChatDto{Name: "Lorem", Id: 19},
		ChatDto{Name: "Impsum", Id: 20},
		ChatDto{Name: "Dolor", Id: 21},
		ChatDto{Name: "With collegues", Id: 22},
		ChatDto{Name: "Lorem", Id: 23},
		ChatDto{Name: "Impsum", Id: 24},
		ChatDto{Name: "Dolor", Id: 25},
		ChatDto{Name: "With collegues", Id: 26},
		ChatDto{Name: "Lorem", Id: 27},
		ChatDto{Name: "Impsum", Id: 28},
		ChatDto{Name: "Dolor", Id: 29},
		ChatDto{Name: "With collegues", Id: 30},
		ChatDto{Name: "Lorem", Id: 31},
		ChatDto{Name: "Impsum", Id: 32},
		ChatDto{Name: "Dolor", Id: 33},
		ChatDto{Name: "With collegues", Id: 34},
		ChatDto{Name: "Lorem", Id: 35},
		ChatDto{Name: "Impsum", Id: 36},
		ChatDto{Name: "Dolor", Id: 37},
		ChatDto{Name: "With collegues", Id: 38},
	}
	return c.JSON(200, usrs)
}

func configureStaticMiddleware() staticMiddleware {
	box := rice.MustFindBox("static").HTTPBox()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqUrl := c.Request().RequestURI
			if reqUrl == "/" || reqUrl == "/index.html" || reqUrl == "/favicon.ico" || strings.HasPrefix(reqUrl, "/build") || strings.HasPrefix(reqUrl, "/assets") {
				http.FileServer(box).
					ServeHTTP(c.Response().Writer, c.Request())
				return nil
			} else {
				return next(c)
			}
		}
	}
}

// rely on viper import and it's configured by
func runEcho(e *echo.Echo) {
	address := viper.GetString("server.address")

	Logger.Info("Starting server...")
	// Start server in another goroutine
	go func() {
		if err := e.Start(address); err != nil {
			Logger.Infof("server shut down: %v", err)
		}
	}()
	Logger.Info("Server started. Waiting for interrupt signal 2 (Ctrl+C)")
}
