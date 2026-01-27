# Redis/Upstash Setup for Gist

This guide explains how to configure Gist to use Redis (specifically Upstash) as its storage backend instead of SQLite.

## Prerequisites

1. **Upstash Account**: Sign up at [Upstash](https://upstash.com/) if you haven't already
2. **Redis Database**: Create a new Redis database in your Upstash console
3. **Environment Variables**: Get your database URL and token from the Upstash dashboard

## Configuration

### Environment Variables

Set the following environment variables to enable Redis storage:

```bash
# Required: Storage backend type
NEXT_PUBLIC_STORAGE_TYPE=redis

# For Redis URL format:
UPSTASH_URL=redis://default:********@*******.upstash.io:****

# Alternative Upstash HTTPS format (requires UPSTASH_TOKEN):
# UPSTASH_URL=https://your-database-id.upstash.io
# UPSTASH_TOKEN=your-database-token-here

# Optional: Other existing Gist configuration
GIST_ADDR=:8080
GIST_DATA_DIR=./data
GIST_LOG_LEVEL=info
```

### Docker Deployment

When deploying with Docker, you can pass these environment variables:

```bash
docker run -d \
  -p 8080:8080 \
  -e NEXT_PUBLIC_STORAGE_TYPE=upstash \
  -e UPSTASH_URL=https://your-database-id.upstash.io \
  -e UPSTASH_TOKEN=your-database-token-here \
  -e GIST_LOG_LEVEL=info \
  your-gist-image
```

### Docker Compose

```yaml
version: '3.8'
services:
  gist:
    image: your-gist-image
    ports:
      - "8080:8080"
    environment:
      - NEXT_PUBLIC_STORAGE_TYPE=upstash
      - UPSTASH_URL=https://your-database-id.upstash.io
      - UPSTASH_TOKEN=your-database-token-here
      - GIST_LOG_LEVEL=info
```

## Data Structure in Redis

When using Redis as the storage backend, Gist organizes data using the following key patterns:

### Feed Data
- `feed:{id}` - Feed object (JSON serialized)
- `feed:url:{url}` - URL to feed ID mapping for lookups

### Entry Data  
- `entry:{id}` - Entry object (JSON serialized)
- `feed:{feed_id}:entries` - Set of entry IDs for a feed

### Folder Data
- `folder:{id}` - Folder object (JSON serialized)

### Settings
- `setting:{key}` - Application settings

### AI Cache
- `ai:summary:{entry_id}:{language}` - AI summary cache (30-day TTL)
- `ai:translation:{entry_id}:{language}` - AI translation cache (30-day TTL)

### Rate Limiting
- `ratelimit:{host}` - Domain rate limit configuration

## Migration from SQLite

**Important**: Currently, Gist does not provide automatic migration from SQLite to Redis. The Redis storage backend is designed for cloud deployments where you want to start fresh.

If you need to migrate existing data:
1. Export your data from SQLite using the existing database
2. Write a custom migration script using the storage abstraction
3. Import the data into Redis

## Development Setup

For local development with Redis:

```bash
# Using a local Redis instance
docker run -d -p 6379:6379 redis:alpine

# Set environment variables
export NEXT_PUBLIC_STORAGE_TYPE=redis
export UPSTASH_URL=redis://default:@localhost:6379
export UPSTASH_TOKEN=

# Run Gist
cd backend
go run ./cmd/server
```

For Upstash Redis:

```bash
# Set environment variables for Upstash
export NEXT_PUBLIC_STORAGE_TYPE=redis
export UPSTASH_URL=redis://default:********@*******.upstash.io:****

# Run Gist
cd backend
go run ./cmd/server
```

## Testing

To run Redis integration tests:

```bash
# Set your Upstash credentials
export UPSTASH_URL=https://your-database-id.upstash.io
export UPSTASH_TOKEN=your-database-token-here

# Run tests
cd backend
make test
```

## Troubleshooting

### Connection Issues

1. **Verify your Upstash URL format**: Should be `https://{database-id}.upstash.io`
2. **Check your token**: Ensure you're using the correct REST token from Upstash
3. **Network connectivity**: Make sure your deployment can reach Upstash endpoints

### Performance

- Redis operations are generally faster than SQLite for read operations
- The current implementation uses basic Redis commands - consider RedisJSON for better JSON handling
- AI cache has a 30-day TTL to prevent unlimited growth

### Data Persistence

- Upstash provides durable storage with persistence
- Data is replicated across multiple nodes for high availability
- Consider your Upstash plan's retention and memory limits

## Limitations

1. **Not all features implemented**: Some advanced SQL features like complex joins are not available
2. **Search functionality**: Full-text search currently requires a separate search service
3. **Backward compatibility**: Existing SQLite-based installations need manual migration

## Future Improvements

- RedisJSON support for better JSON document handling
- Redis Search integration for full-text search
- Automatic data migration tools
- Redis Streams for real-time updates
- Lua scripting for complex operations

For more information about Upstash Redis, visit the [Upstash Documentation](https://docs.upstash.com/redis).