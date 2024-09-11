package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/client/v3/snapshot"
	"go.uber.org/zap"
)

// BackupReconciler handles the backup process
type BackupReconciler struct {
	etcdClient *clientv3.Client
	etcdConfig clientv3.Config // Store the config for snapshot
}

// initEtcdClient initializes the etcd client
func initEtcdClient() (*clientv3.Client, clientv3.Config, error) {
	endpoints := []string{"http://127.0.0.1:2379"} // Your etcd endpoint(s)
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, config, fmt.Errorf("failed to connect to etcd: %w", err)
	}
	return client, config, nil
}

// takeEtcdSnapshot uses etcd's snapshot API to save a backup
func (r *BackupReconciler) takeEtcdSnapshot(ctx context.Context, snapshotFilePath string) error {
	log.Println("Starting etcd snapshot backup...")

	// Acquire lock to ensure only one backup happens at a time
	session, err := concurrency.NewSession(r.etcdClient)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, "/etcd-backup-lock")
	if err := mutex.Lock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer mutex.Unlock(ctx)

	// Create a logger for snapshot
	logger, _ := zap.NewProduction()

	// Call the snapshot API
	if err := snapshot.Save(ctx, logger, r.etcdConfig, snapshotFilePath); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	log.Printf("Snapshot saved successfully at %s\n", snapshotFilePath)
	return nil
}

func main() {
	// Initialize etcd client
	etcdClient, etcdConfig, err := initEtcdClient()
	if err != nil {
		log.Fatalf("Error initializing etcd client: %v", err)
	}
	defer etcdClient.Close()

	reconciler := &BackupReconciler{etcdClient: etcdClient, etcdConfig: etcdConfig}

	// Run snapshot process
	snapshotFile := fmt.Sprintf("/backup/etcd_snapshot_%s.db", time.Now().Format("2006-01-02_15-04-05"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := reconciler.takeEtcdSnapshot(ctx, snapshotFile); err != nil {
		log.Fatalf("Error taking etcd snapshot: %v", err)
	}
}
