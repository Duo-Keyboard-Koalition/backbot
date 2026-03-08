# Task 005: Tailscale ACL Integration

## Priority
**P1** - High

## Status
⏳ Pending

## Objective
Integrate Tailscale ACLs (Access Control Lists) with Taila2a for fine-grained authorization of agent capabilities and topic access.

## Background
Tailscale ACLs provide network-level access control. By integrating with Tailscale's ACL system, Taila2a can enforce authorization at the application level based on user identity and device tags.

## Requirements

### ACL Integration
- [ ] Read Tailscale ACL policy from control plane
- [ ] Map Tailscale users/devices to Taila2a roles
- [ ] Enforce topic-level access control
- [ ] Enforce capability-based access
- [ ] Audit logging for access decisions

### Tag-Based Authorization
- [ ] Define tags for agent roles
- [ ] Map tags to capabilities
- [ ] Validate agent tags on connection
- [ ] Reject unauthorized agents

### User-Based Authorization
- [ ] Identify users via Tailscale identity
- [ ] Map users to Taila2a permissions
- [ ] Support user groups
- [ ] Per-user rate limits

### API
- [ ] `CheckTopicAccess(user, topic, action) -> bool`
- [ ] `CheckCapability(user, capability) -> bool`
- [ ] `GetUserPermissions(user) -> []Permission`
- [ ] `GetEffectiveACL(node) -> ACL`

## Technical Specification

### Tag Definitions
```json
{
  "tag:taila2a:agent": {
    "description": "General agent capability",
    "capabilities": ["chat", "command"]
  },
  "tag:taila2a:producer": {
    "description": "Message producer",
    "capabilities": ["publish"]
  },
  "tag:taila2a:consumer": {
    "description": "Message consumer",
    "capabilities": ["subscribe"]
  },
  "tag:taila2a:admin": {
    "description": "Administrative access",
    "capabilities": ["admin", "publish", "subscribe", "manage_topics"]
  },
  "tag:tailfs:sender": {
    "description": "Can send files",
    "capabilities": ["file_send"]
  },
  "tag:tailfs:receiver": {
    "description": "Can receive files",
    "capabilities": ["file_receive"]
  }
}
```

### ACL Policy Format
```json
{
  "groups": {
    "group:developers": ["alice@example.com", "bob@example.com"],
    "group:admins": ["admin@example.com"]
  },
  "hosts": {
    "prod-agents": ["tag:taila2a:agent"],
    "dev-agents": ["tag:taila2a:dev"]
  },
  "acls": [
    {
      "action": "accept",
      "src": ["group:developers"],
      "dst": ["prod-agents:8001"],
      "proto": "tcp"
    },
    {
      "action": "accept",
      "src": ["tag:taila2a:producer"],
      "dst": ["tag:taila2a:agent:8001"],
      "proto": "tcp"
    }
  ],
  "taila2a": {
    "topic_access": [
      {
        "topics": ["agent.requests"],
        "actions": ["publish"],
        "src": ["tag:taila2a:producer"]
      },
      {
        "topics": ["agent.responses"],
        "actions": ["subscribe"],
        "src": ["tag:taila2a:consumer"]
      }
    ],
    "capability_access": [
      {
        "capabilities": ["file_send", "file_receive"],
        "src": ["tag:tailfs:*"]
      }
    ]
  }
}
```

### ACL Checker Interface
```go
type ACLChecker interface {
    // CheckTopicAccess checks if src can perform action on topic
    CheckTopicAccess(src identity.Identity, topic string, action string) bool
    
    // CheckCapability checks if src has capability
    CheckCapability(src identity.Identity, capability string) bool
    
    // GetUserPermissions returns all permissions for user
    GetUserPermissions(userID string) []Permission
    
    // ValidateAgentTags validates agent's tags
    ValidateAgentTags(nodeID string, requiredTags []string) bool
    
    // GetACL returns the current ACL policy
    GetACL() *ACLPolicy
}
```

## Acceptance Criteria

### Functional
- [ ] ACL policies are loaded from Tailscale
- [ ] Topic access is enforced
- [ ] Capability access is enforced
- [ ] Tag-based authorization works
- [ ] User-based authorization works
- [ ] Audit logs are generated

### Non-Functional
- [ ] ACL check latency < 1ms
- [ ] Policy updates applied within 5 seconds
- [ ] No unauthorized access possible
- [ ] Graceful degradation if Tailscale unavailable

## Testing Requirements

### Unit Tests
- [ ] Test ACL parsing
- [ ] Test tag matching
- [ ] Test user group resolution
- [ ] Test permission checks

### Integration Tests
- [ ] Test with real Tailscale ACLs
- [ ] Test policy updates
- [ ] Test unauthorized access rejection

### Security Tests
- [ ] Test privilege escalation attempts
- [ ] Test tag spoofing prevention
- [ ] Test bypass attempts

## Implementation Steps

1. **Phase 1: ACL Loading** (2 days)
   - Connect to Tailscale control plane
   - Parse ACL policy
   - Cache policy locally

2. **Phase 2: Tag-Based Auth** (2 days)
   - Implement tag validation
   - Map tags to capabilities
   - Enforce on connections

3. **Phase 3: Topic ACLs** (2 days)
   - Implement topic access checks
   - Integrate with event bus
   - Add audit logging

4. **Phase 4: User-Based Auth** (2 days)
   - Resolve user identities
   - Implement user group checks
   - Add per-user rate limits

5. **Phase 5: Testing** (2 days)
   - Security testing
   - Integration testing
   - Documentation

## Files to Create/Modify

### Create
- `taila2a/internal/acl/acl.go` - ACL checker
- `taila2a/internal/acl/policy.go` - Policy parsing
- `taila2a/internal/acl/tags.go` - Tag management
- `taila2a/internal/acl/identity.go` - Identity resolution
- `taila2a/internal/acl/audit.go` - Audit logging
- `taila2a/internal/acl/acl_test.go` - Tests

### Modify
- `taila2a/internal/controllers/controller.go` - Add ACL checks
- `taila2a/internal/controllers/phonebook_controller.go` - Filter by ACL
- `taila2a/cmd/taila2a/main.go` - Initialize ACL checker

## References
- [Tailscale ACLs](https://tailscale.com/kb/1018/acls/)
- [Tailscale API](https://tailscale.com/kb/1101/api/)
- [A2A Protocol Security](../eng_nbk/A2A_PROTOCOL.md#security-model)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
