package cron

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"
)

type Cron struct {
	client   *cron.Cron
	schedule string
	jobName  string
}

func New(schedule, jobName string) *Cron {
	return &Cron{
		client:   cron.New(),
		schedule: schedule,
		jobName:  jobName,
	}
}

func (c *Cron) Setup(job func(context.Context)) (context.CancelFunc, error) {
	cronCtx, cancelCron := context.WithCancel(context.Background())

	err := c.addJob(cronCtx, job)
	if err != nil {
		return cancelCron, err
	}
	c.start()

	return cancelCron, nil
}

func (c *Cron) addJob(ctx context.Context, job func(context.Context)) error {
	_, err := c.client.AddFunc(c.schedule, func() {
		job(ctx)
	})
	return err
}

func (c *Cron) start() {
	c.client.Start()
	slog.Info("cron scheduler started",
		slog.String("schedule", c.schedule),
		slog.String("job", c.jobName))
}

func (c *Cron) Stop() context.Context {
	return c.client.Stop()
}
