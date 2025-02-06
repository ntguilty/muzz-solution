# Muzz Backend Engineer Exercise - Solution Overview

## Architecture Decisions

### Database Schema
Decided to use a single table solution with appropriate index. The main reason was that the service in my opinion is read-heavy and we needed to optimize for quick data retrieval.

Considered partitioning and other advanced PostgreSQL features but decided against them as it would be over-engineering for a recruitment task. The chosen solution with simple indexes provides good performance while maintaining simplicity.

### Caching Strategy
Implemented Redis caching for both listings and counters because:
- The data is read-heavy (users frequently check who liked them)
- The data is eventually consistent (small delay in seeing new likes is acceptable)
- Lists and counts are computationally expensive, especially with large datasets

### Design Decisions
- Cursor-based pagination using timestamps instead of offset-based
    - Better performance with large datasets
    - Consistent results even when new likes are added
- NOT EXISTS instead of JOINs for mutual likes check
    - Better performance as it can use indexes effectively
    - Simpler query plan
- Base64 encoded pagination tokens
    - Clean response

### Trade-offs
- Sacrificed some write performance (due to indexes) to gain better read performance
- Accepted eventual consistency in cache for better performance
- Chose simpler implementation over more complex optimizations that might be needed in production

## What Could Be Added in Production
- Proper monitoring and metrics
- Rate limiting
- More sophisticated caching strategies
- Better error handling and recovery mechanisms
- Database optimizations based on metrics e.g. table partitioning or read replicas