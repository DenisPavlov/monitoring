// Package storage provides comprehensive metrics storage solutions including:
// - In-memory storage with thread-safe operations
// - File-based persistence with JSON serialization
// - PostgreSQL database storage with retry logic and connection resilience
// - Unified interface for consistent API across different storage implementations
//
// The package supports various storage backends with capabilities for:
// - Atomic operations and transaction support
// - Concurrent access with proper synchronization
// - Automatic retry mechanisms for transient failures
// - Metrics persistence and recovery
package storage
