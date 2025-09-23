package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"vk-normalizer/internal/config"
	"vk-normalizer/internal/logger"
	"vk-normalizer/internal/messaging"
	"vk-normalizer/internal/s3"
	"vk-normalizer/internal/vk"
	"vk-normalizer/internal/vknorm"
)

func main() {
	// init logger first to use it everywhere
	lg, cleanup := logger.NewZapLogger()
	defer cleanup()
	lg.Info("🚀 Starting vk-normalizer…")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		lg.Error("❌ Failed to load config: %v", err)
		os.Exit(1)
	}

	uploader, err := s3.NewUploader(s3.Option{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		UseSSL:    cfg.S3.UseSSL,
	})
	if err != nil {
		lg.Error("❌ Failed to create S3 uploader: %v", err)
		os.Exit(1)
	}

	downloader := vk.NewDownloader()

	producer, err := messaging.NewKafkaProducer(messaging.Option{
		Logger:       lg,
		Broker:       cfg.Kafka.BootstrapServersValue,
		SaslUsername: cfg.Kafka.SaslUsername,
		SaslPassword: cfg.Kafka.SaslPassword,
		ClientID:     cfg.Kafka.ClientID,
	})
	if err != nil {
		lg.Error("❌ Failed to create Kafka producer: %v", err)
		os.Exit(1)
	}

	parser := vk.NewVkMessageParser()

	normalizer := vknorm.NewVkNormalizer(vknorm.Option{
		KafkaTopic: cfg.Kafka.NormalizerTopicName,
		Uploader:   uploader,
		Parser:     parser,
		Downloader: downloader,
		Producer:   producer,
	})

	consumer, err := messaging.NewKafkaConsumer(messaging.ConsumerOption{
		Logger:       lg,
		Broker:       cfg.Kafka.BootstrapServersValue,
		GroupID:      cfg.Kafka.GroupID,
		Topics:       []string{cfg.Kafka.VkMessTopicName},
		Handler:      normalizer,
		SaslUsername: cfg.Kafka.SaslUsername,
		SaslPassword: cfg.Kafka.SaslPassword,
		ClientID:     cfg.Kafka.ClientID,
	})
	if err != nil {
		lg.Error("❌ Failed to create consumer: %v", err)
		os.Exit(1)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	if err := consumer.Start(ctx); err != nil {
		lg.Error("❌ Consumer error: %v", err)
		os.Exit(1)
	}
}
