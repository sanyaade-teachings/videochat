package tasks

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/nkonev/dcron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	. "nkonev.name/storage/logger"
	"nkonev.name/storage/s3"
	"nkonev.name/storage/services"
	"nkonev.name/storage/utils"
	"time"
)

type ActualizeGeneratedFilesTask struct {
	dcron.Job
}

func ActualizeGeneratedFilesScheduler(
	lgr *log.Logger,
	service *ActualizeGeneratedFilesService,
) *ActualizeGeneratedFilesTask {
	const key = "actualizeGeneratedFilesTask"
	var str = viper.GetString("schedulers." + key + ".cron")
	lgr.Infof("Created ActualizeGeneratedFilesScheduler with cron %v", str)

	job := dcron.NewJob(key, str, func(ctx context.Context) error {
		service.doJob()
		return nil
	})

	return &ActualizeGeneratedFilesTask{job}
}

type ActualizeGeneratedFilesService struct {
	minioClient        *s3.InternalMinioClient
	minioBucketsConfig *utils.MinioConfig
	previewService     *services.PreviewService
	tracer             trace.Tracer
	redisInfoService   *services.RedisInfoService
	convertingService  *services.ConvertingService
	lgr                *log.Logger
}

func (srv *ActualizeGeneratedFilesService) doJob() {
	ctx, span := srv.tracer.Start(context.Background(), "scheduler.ActualizeGeneratedFiles")
	defer span.End()
	filenameChatPrefix := "chat/"
	srv.processFiles(ctx, filenameChatPrefix)
}

func (srv *ActualizeGeneratedFilesService) processFiles(c context.Context, filenameChatPrefix string) {
	GetLogEntry(c, srv.lgr).Infof("Starting actualize generated files job")

	// create preview for files if need
	// and create _converted.webm
	GetLogEntry(c, srv.lgr).Infof("Checking for missing previews and converted")
	var fileObjects <-chan minio.ObjectInfo = srv.minioClient.ListObjects(c, srv.minioBucketsConfig.Files, minio.ListObjectsOptions{
		Prefix:       filenameChatPrefix,
		Recursive:    true,
		WithMetadata: true,
	})
	for fileOjInfo := range fileObjects {
		// here in minio 'chat/108/'
		GetLogEntry(c, srv.lgr).Debugf("Start processing minio key '%v'", fileOjInfo.Key)
		if utils.IsVideo(fileOjInfo.Key) {
			// preview
			previewToCheck := utils.SetVideoPreviewExtension(fileOjInfo.Key)

			previewExists, _, err := srv.minioClient.FileExists(c, srv.minioBucketsConfig.FilesPreview, previewToCheck)
			if err != nil {
				GetLogEntry(c, srv.lgr).Errorf("Unable to check existence for %v: %v", previewToCheck, err)
				continue
			}
			if !previewExists {
				GetLogEntry(c, srv.lgr).Infof("Create missed preview %v for %v", previewToCheck, fileOjInfo.Key)
				srv.previewService.CreatePreview(c, fileOjInfo.Key)
			}

			// _converted.webm
			_, _, _, isMessageRecording, err := services.DeserializeMetadata(fileOjInfo.UserMetadata, true)
			if err != nil {
				GetLogEntry(c, srv.lgr).Errorf("Unable to convert metadata for key %v: %v", fileOjInfo.Key, err)
				continue
			}
			isConverting, err := srv.redisInfoService.GetOriginalConverting(c, fileOjInfo.Key)
			if err != nil {
				GetLogEntry(c, srv.lgr).Errorf("Unable to isConverting for key %v from redis: %v", fileOjInfo.Key, err)
				continue
			}

			keyOfConverted := utils.GetKeyForConverted(fileOjInfo.Key)
			convertedExists, _, err := srv.minioClient.FileExists(c, srv.minioBucketsConfig.Files, keyOfConverted)
			if err != nil {
				GetLogEntry(c, srv.lgr).Errorf("Unable to check existence for %v: %v", keyOfConverted, err)
				continue
			}
			if !convertedExists && utils.IsVideo(fileOjInfo.Key) && utils.NullableToBoolean(isMessageRecording) && !utils.IsConverted(fileOjInfo.Key) && !isConverting {
				GetLogEntry(c, srv.lgr).Infof("Create missed converted for %v", fileOjInfo.Key)
				srv.convertingService.Convert(c, fileOjInfo.Key)
			}
		} else if utils.IsImage(fileOjInfo.Key) {
			previewToCheck := utils.SetImagePreviewExtension(fileOjInfo.Key)
			exists, _, err := srv.minioClient.FileExists(c, srv.minioBucketsConfig.FilesPreview, previewToCheck)
			if err != nil {
				GetLogEntry(c, srv.lgr).Errorf("Unable to check existence for %v: %v", previewToCheck, err)
				continue
			}
			if !exists {
				GetLogEntry(c, srv.lgr).Infof("Create preview for missing %v", fileOjInfo.Key)
				srv.previewService.CreatePreview(c, fileOjInfo.Key)
			}
		}

	}
	GetLogEntry(c, srv.lgr).Infof("Checking for missing previews and converted finished")

	// remove previews of removed files
	GetLogEntry(c, srv.lgr).Infof("Checking for excess previews")
	var previewObjects <-chan minio.ObjectInfo = srv.minioClient.ListObjects(c, srv.minioBucketsConfig.FilesPreview, minio.ListObjectsOptions{
		Prefix:       filenameChatPrefix,
		Recursive:    true,
		WithMetadata: true,
	})
	for previewOjInfo := range previewObjects {
		GetLogEntry(c, srv.lgr).Debugf("Start processing minio key '%v'", previewOjInfo.Key)
		originalKey, err := services.GetOriginalKeyFromMetadata(previewOjInfo.UserMetadata, true)
		if err != nil {
			GetLogEntry(c, srv.lgr).Errorf("Error during getting original key %v", err)
			continue
		}
		exists, _, err := srv.minioClient.FileExists(c, srv.minioBucketsConfig.Files, originalKey)
		if err != nil {
			GetLogEntry(c, srv.lgr).Errorf("Unable to get exists for %v: %v", originalKey, err)
			continue
		}

		maxConvertingDuration := viper.GetDuration("converting.maxDuration")
		if !exists {
			GetLogEntry(c, srv.lgr).Infof("Key %v is not found, deciding whether to remove the preview %v", originalKey, previewOjInfo.Key)
			if utils.IsConverted(originalKey) && previewOjInfo.LastModified.Add(maxConvertingDuration).After(time.Now().UTC()) {
				GetLogEntry(c, srv.lgr).Infof("Age of the converted preview %v for key %v is lesser than %v, skipping deletion", previewOjInfo.Key, originalKey, maxConvertingDuration)
				continue
			} else {
				GetLogEntry(c, srv.lgr).Infof("Will remove preview for %v", originalKey)
				err := srv.minioClient.RemoveObject(c, srv.minioBucketsConfig.FilesPreview, previewOjInfo.Key, minio.RemoveObjectOptions{})
				if err != nil {
					GetLogEntry(c, srv.lgr).Errorf("Error during removing preview key %v", err)
					continue
				}
			}
		}
	}
	GetLogEntry(c, srv.lgr).Infof("Checking for excess previews finished")

	GetLogEntry(c, srv.lgr).Infof("End of generated files job")
}

func NewActualizeGeneratedFilesService(lgr *log.Logger, minioClient *s3.InternalMinioClient, minioBucketsConfig *utils.MinioConfig, previewService *services.PreviewService, redisInfoService *services.RedisInfoService, convertingService *services.ConvertingService) *ActualizeGeneratedFilesService {
	trcr := otel.Tracer("scheduler/actualize-generated-files")
	return &ActualizeGeneratedFilesService{
		lgr:                lgr,
		minioClient:        minioClient,
		minioBucketsConfig: minioBucketsConfig,
		previewService:     previewService,
		tracer:             trcr,
		redisInfoService:   redisInfoService,
		convertingService:  convertingService,
	}
}
