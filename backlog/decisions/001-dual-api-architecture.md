# Decision Record: Dual API Architecture

**Date**: 2025-07-12  
**Status**: Accepted  
**Deciders**: Development Team  

## Context

GitHub provides two primary APIs for repository management:
- **REST v3 API**: Mature, comprehensive for certain operations
- **GraphQL v4 API**: Modern, efficient, but feature gaps exist

We needed to decide on an API strategy that would provide comprehensive coverage of GitHub's repository management features.

## Decision

We will maintain a **dual API architecture** that leverages both GitHub's REST v3 and GraphQL v4 APIs, using each API for its strengths.

### API Usage Strategy

| Feature Category | Primary API | Fallback API | Rationale |
|------------------|-------------|--------------|-----------|
| Team Permissions | REST v3 | None | Only available in REST v3 |
| Repository Settings | GraphQL v4 | REST v3 | Better batch operations in GraphQL |
| Branch Protection | GraphQL v4 | REST v3 | More comprehensive in GraphQL |
| Issue Labels | REST v3 | None | More mature implementation |
| Repository Archiving | GraphQL v4 | None | Only available in GraphQL |
| Merge Strategies | REST v3 | None | Better coverage in REST |

## Consequences

### Positive
- **Comprehensive Feature Coverage**: Access to all GitHub repository management features
- **Performance Optimization**: Use most efficient API for each operation
- **Future-Proofing**: Can adopt new features as they become available in either API
- **Reliability**: Fallback options for critical operations

### Negative
- **Increased Complexity**: Need to maintain two API clients
- **Additional Testing**: Both API paths require comprehensive testing
- **Dependency Management**: More external dependencies to manage
- **Learning Curve**: Developers need to understand both APIs

### Neutral
- **Code Organization**: Clear separation of concerns between API clients
- **Error Handling**: Unified error handling across both APIs

## Implementation Details

### Client Structure
```go
type GitHubClients struct {
    RestClient    *github.Client     // REST v3 operations
    GraphQLClient *githubv4.Client   // GraphQL v4 operations
}
```

### Error Handling
- Unified error types across both APIs
- Graceful fallback from GraphQL to REST where applicable
- Clear error messages indicating which API failed

### Testing Strategy
- Mock both API clients independently
- Integration tests for critical fallback scenarios
- Performance comparisons between API approaches

## Alternatives Considered

### REST v3 Only
- **Pros**: Simpler architecture, mature API
- **Cons**: Missing modern features (advanced branch protection, efficient archiving)
- **Decision**: Rejected due to feature limitations

### GraphQL v4 Only
- **Pros**: Modern, efficient, single API
- **Cons**: Missing critical features (team permissions, comprehensive labels)
- **Decision**: Rejected due to feature gaps

### API Abstraction Layer
- **Pros**: Hide API differences from business logic
- **Cons**: Additional complexity, potential performance overhead
- **Decision**: Deferred - may implement in future if complexity becomes unmanageable

## Review Schedule

This decision will be reviewed:
- **Quarterly**: Check for new GitHub API capabilities
- **Before Major Releases**: Evaluate if architecture still serves project needs
- **When API Deprecations**: GitHub announces deprecation of either API

## Related Decisions

- **Configuration Strategy**: YAML-based configuration to abstract API differences
- **Error Handling**: Unified error types across both APIs
- **Testing Strategy**: Comprehensive testing of both API paths