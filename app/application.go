package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/SriHemnath/credit_card_processing/config"
	"github.com/SriHemnath/credit_card_processing/streams/kafka"
	"github.com/SriHemnath/credit_card_processing/utils/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func StartApplication() {
	app := fiber.New()

	mapURLs(app)

	//load config from yml
	cfg := loadConfig()
	var wg sync.WaitGroup

	//cancellable context, cancel must be called to stop event processing and graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		signs := make(chan os.Signal)
		signal.Notify(signs, os.Interrupt, syscall.SIGTERM)
		select {
		case sig := <-signs:
			logger.Info(fmt.Sprintf("received signal %s, shutting down", sig))
			app.Shutdown()
		case <-ctx.Done():
		}
		cancel()
	}()

	//http server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting server in :8000....")
		err := app.Listen(":8000")
		if err != nil && err != context.Canceled {
			logger.Error("server error: ", err)
		}
		cancel()
	}()
	logger.Info("Server started successfully....")

	//consumer setup
	consumer, err := createConsumer(cfg)
	if err != nil {
		log.Fatalf("Error while creating consumer: %v", err)
	}

	//kafka consumer
	wg.Add(1)
	go func() {
		defer wg.Wait()
		consumeErr := consumer.Consume(ctx)
		if consumeErr != nil {
			logger.Error("kafka consumer error: ", consumeErr)
		}
		consumer.Close()
		cancel()
	}()
	// wait for go singal listener
	wg.Wait()

	//may block the entire program
	runtime.GC()
	logger.Info("Graceful shutdown")

}

func loadConfig() (cfg config.Config) {
	viper.SetConfigFile("./.conf/e0.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("unable to read config using viper")
	}

	runtime.GC()
	//var cfg config.Config

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("unable to unmarshal config using viper")
	}

	log.Printf("%+v\n", cfg)
	return
}

func createConsumer(cfg config.Config) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(
		cfg.Consumer.Brokers,
		cfg.Consumer.Group,
		cfg.Consumer.Topic,
		cfg.Consumer.CommitInterval,
	)
	return consumer, err
}
