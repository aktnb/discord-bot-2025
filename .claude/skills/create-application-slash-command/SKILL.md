---
name: create-application-slash-command
description: Creates Application Slash Command definition and handler following clean architecture
---

## Overview

This guide explains how to add new Discord Application Slash Commands to the bot following clean architecture principles. The implementation is organized into layers with clear separation of concerns.

## Architecture

This project follows Clean Architecture (Hexagonal Architecture) with the following layers:

```
┌─────────────────────────────────────────────────────────┐
│  Infrastructure Layer (Discord, External APIs)          │
│  ├─ discord/commands/*/command.go (CommandDefinition)  │
│  ├─ discord/commands/*/command.go (CommandHandler)     │
│  └─ */client.go (External API clients)                 │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Application Layer (Use Cases)                          │
│  └─ application/*/service.go                            │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Domain Layer (Business Logic)                          │
│  ├─ domain/*/model.go (Entities)                        │
│  ├─ domain/*/repository.go (Port interfaces)            │
│  └─ domain/*/errors.go (Domain errors)                  │
└─────────────────────────────────────────────────────────┘
```

**Key Principles:**
- Outer layers depend on inner layers
- Inner layers are independent of outer layers
- Domain layer has no external dependencies
- Infrastructure implements ports defined by domain

## Implementation Patterns

### Pattern A: Simple Command (like `ping`)
Use when:
- No external dependencies
- Quick, synchronous response
- Simple business logic

**Layers needed:**
- Application layer (service)
- Infrastructure layer (command definition & handler)

### Pattern B: External API Integration (like `cat`)
Use when:
- External API calls required
- Async operations (>3 seconds)
- Complex business logic

**Layers needed:**
- Domain layer (entities, repository interface, errors)
- Application layer (service)
- Infrastructure layer (API client, command definition & handler)

## Step-by-Step Guide

### Pattern A: Simple Command

#### 1. Create Application Service
**File:** `internal/application/{command}/service.go`

```go
package {command}

import "context"

type Service struct{}

func New{Command}Service() *Service {
	return &Service{}
}

func (s *Service) {Action}(ctx context.Context) (string, error) {
	// Business logic here
	return "Response message", nil
}
```

#### 2. Create Command Definition & Handler
**File:** `internal/infrastructure/discord/commands/{command}/command.go`

```go
package {command}

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/{command}"
	"github.com/bwmarrin/discordgo"
)

type {Command}CommandDefinition struct{}

func New{Command}CommandDefinition() *{Command}CommandDefinition {
	return &{Command}CommandDefinition{}
}

func (c *{Command}CommandDefinition) Name() string {
	return "{command}"
}

func (c *{Command}CommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Command description in Japanese",
	}
}

type {Command}CommandHandler struct {
	service *{command}.Service
}

func New{Command}CommandHandler(service *{command}.Service) *{Command}CommandHandler {
	return &{Command}CommandHandler{
		service: service,
	}
}

func (h *{Command}CommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	response, err := h.service.{Action}(ctx)
	if err != nil {
		log.Printf("Error handling {command} command: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Printf("Error responding to {command}: %v", err)
		return err
	}

	return nil
}
```

#### 3. Register in main.go
**File:** `cmd/bot/main.go`

```go
// Import
import (
	"{command}app" "github.com/aktnb/discord-bot-go/internal/application/{command}"
	{command}cmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/{command}"
)

// Registration (after command registry creation)
{command}Service := {command}app.New{Command}Service()
{command}Def := {command}cmd.New{Command}CommandDefinition()
{command}Handler := {command}cmd.New{Command}CommandHandler({command}Service)

registry.Register({command}Def, {command}Handler)
```

### Pattern B: External API Integration

#### 1. Create Domain Layer

**File:** `internal/domain/{command}/model.go`
```go
package {command}

type {Entity} struct {
	// Entity fields
	ID   string
	Data string
}
```

**File:** `internal/domain/{command}/repository.go`
```go
package {command}

import "context"

type {Entity}Repository interface {
	Fetch{Entity}(ctx context.Context) (*{Entity}, error)
}
```

**File:** `internal/domain/{command}/errors.go`
```go
package {command}

import "errors"

var (
	Err{Entity}NotFound   = errors.New("{entity} not found")
	ErrAPIUnavailable     = errors.New("external API is unavailable")
	ErrInvalidResponse    = errors.New("invalid API response")
)
```

#### 2. Create Application Service

**File:** `internal/application/{command}/service.go`
```go
package {command}

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/{command}"
)

type Service struct {
	repo {command}.{Entity}Repository
}

func New{Command}Service(repo {command}.{Entity}Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get{Entity}(ctx context.Context) (*{command}.{Entity}, error) {
	entity, err := s.repo.Fetch{Entity}(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
```

#### 3. Create External API Client

**File:** `internal/infrastructure/{api}/client.go`
```go
package {api}

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/{command}"
)

const (
	baseURL        = "https://api.example.com/endpoint"
	requestTimeout = 10 * time.Second
)

type apiResponse struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type {API}Client struct {
	httpClient *http.Client
}

func New{API}Client() *{API}Client {
	return &{API}Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *{API}Client) Fetch{Entity}(ctx context.Context) (*{command}.{Entity}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, {command}.ErrAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, {command}.ErrAPIUnavailable
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, {command}.ErrInvalidResponse
	}

	return &{command}.{Entity}{
		ID:   apiResp.ID,
		Data: apiResp.Data,
	}, nil
}
```

#### 4. Create Command with Deferred Response

**File:** `internal/infrastructure/discord/commands/{command}/command.go`
```go
package {command}

import (
	"context"
	"log"

	app{command} "github.com/aktnb/discord-bot-go/internal/application/{command}"
	"github.com/bwmarrin/discordgo"
)

type {Command}CommandDefinition struct{}

func New{Command}CommandDefinition() *{Command}CommandDefinition {
	return &{Command}CommandDefinition{}
}

func (c *{Command}CommandDefinition) Name() string {
	return "{command}"
}

func (c *{Command}CommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Command description in Japanese",
	}
}

type {Command}CommandHandler struct {
	service *app{command}.Service
}

func New{Command}CommandHandler(service *app{command}.Service) *{Command}CommandHandler {
	return &{Command}CommandHandler{
		service: service,
	}
}

func (h *{Command}CommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Defer response for async operations
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	entity, err := h.service.Get{Entity}(ctx)
	if err != nil {
		log.Printf("Error fetching entity: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラーが発生しました。もう一度お試しください。",
		})
		return err
	}

	// Send the result
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: entity.Data,
	})
	if err != nil {
		log.Printf("Error sending response: %v", err)
		return err
	}

	return nil
}
```

#### 5. Register in main.go

**File:** `cmd/bot/main.go`
```go
// Imports
import (
	"{command}app" "github.com/aktnb/discord-bot-go/internal/application/{command}"
	{command}cmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/{command}"
	"{api}" "github.com/aktnb/discord-bot-go/internal/infrastructure/{api}"
)

// Registration
{api}Client := {api}.New{API}Client()
{command}Service := {command}app.New{Command}Service({api}Client)
{command}Def := {command}cmd.New{Command}CommandDefinition()
{command}Handler := {command}cmd.New{Command}CommandHandler({command}Service)

registry.Register({command}Def, {command}Handler)
```

## Technical Considerations

### Discord 3-Second Rule
Discord requires a response within 3 seconds, or the interaction will timeout. For operations that may take longer:

1. Use `InteractionResponseDeferredChannelMessageWithSource` immediately
2. Perform the actual work
3. Send the result via `FollowupMessageCreate`

### Error Handling
- Log all errors with `log.Printf`
- Return user-friendly Japanese error messages
- Define domain-specific errors in `domain/*/errors.go`
- Convert infrastructure errors to domain errors

### Logging Best Practices
- Log command start/completion
- Log errors with context
- Include command name in log messages
- Use structured logging format

## Command Interfaces

All commands must implement these interfaces defined in `internal/infrastructure/discord/commands/interfaces.go`:

```go
type CommandDefinition interface {
	Name() string
	ToDiscordCommand() *discordgo.ApplicationCommand
}

type CommandHandler interface {
	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}
```

## Implementation Checklist

- [ ] Choose appropriate pattern (A or B)
- [ ] Create directory structure for the command
- [ ] Implement domain layer (if Pattern B)
  - [ ] Define entities in `model.go`
  - [ ] Define repository interface in `repository.go`
  - [ ] Define domain errors in `errors.go`
- [ ] Implement application service
  - [ ] Create `service.go` with business logic
  - [ ] Inject dependencies via constructor
- [ ] Implement infrastructure layer
  - [ ] Create external API client (if needed)
  - [ ] Create `CommandDefinition` struct
  - [ ] Create `CommandHandler` struct
  - [ ] Use deferred response if operation may take >3 seconds
- [ ] Register command in `main.go`
  - [ ] Add imports
  - [ ] Create service instance
  - [ ] Create definition and handler instances
  - [ ] Register with command registry
- [ ] Test the command
  - [ ] Run `go build ./cmd/bot/main.go`
  - [ ] Start the bot
  - [ ] Verify command appears in Discord
  - [ ] Test command execution
  - [ ] Test error cases

## Reference Files

**Pattern A (Simple):**
- `internal/application/ping/service.go`
- `internal/infrastructure/discord/commands/ping/command.go`

**Pattern B (External API):**
- `internal/domain/cat/model.go`
- `internal/domain/cat/repository.go`
- `internal/domain/cat/errors.go`
- `internal/application/cat/service.go`
- `internal/infrastructure/catapi/client.go`
- `internal/infrastructure/discord/commands/cat/command.go`

**Registration:**
- `cmd/bot/main.go` (lines 50-66 for examples)

## Directory Structure Example

```
internal/
├── domain/
│   └── {command}/
│       ├── model.go        # Entities (Pattern B only)
│       ├── repository.go   # Port interfaces (Pattern B only)
│       └── errors.go       # Domain errors (Pattern B only)
├── application/
│   └── {command}/
│       └── service.go      # Business logic
└── infrastructure/
    ├── {api}/              # External API client (Pattern B only)
    │   └── client.go
    └── discord/
        └── commands/
            └── {command}/
                └── command.go  # CommandDefinition & CommandHandler
```
