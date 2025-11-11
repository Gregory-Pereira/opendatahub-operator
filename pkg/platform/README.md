# Platform Package

The platform package provides an abstraction layer for platform-specific behavior in the OpenDataHub Operator. It enables different deployment variants (SelfManaged, OpenDataHub, Vanilla) to implement custom logic for upgrades, initialization, runtime operations, and validation.

## Architecture

### Core Interfaces

#### Platform Interface
Defines the contract for platform-specific implementations:
- `Upgrade(ctx context.Context) error` - Platform-specific upgrade logic
- `Init(ctx context.Context) error` - Platform initialization during operator startup
- `Run(ctx context.Context, mgr ctrl.Manager) error` - Platform runtime logic during reconciliation with access to the controller manager
- `Validator() Validator` - Returns platform-specific webhook validators
- `String() string` - Returns the canonical platform display name (implements `fmt.Stringer`)

#### Validator Interface
Provides admission webhook handlers for CR validation:
- `DSCInitializationValidator() admission.Handler` - DSCI validation webhook
- `DataScienceClusterValidator() admission.Handler` - DSC validation webhook

## Platform Variants

### SelfManaged (`pkg/platform/selfmanaged`)
Platform implementation for self-managed OpenDataHub deployments.

### OpenDataHub (`pkg/platform/opendatahub`)
Platform implementation for OpenDataHub community deployments.

### Managed (`pkg/platform/managed`)
Platform implementation for managed OpenDataHub service deployments.

### Vanilla (`pkg/platform/vanilla`)
Platform implementation for vanilla Kubernetes deployments without OpenShift dependencies.

## Usage

### Creating a Platform Instance

```go
import (
    "github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/factory"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// Get Kubernetes client (from manager or elsewhere)
var cli client.Client

// Create platform based on type (read from ODH_PLATFORM_TYPE env var)
platformType := os.Getenv("ODH_PLATFORM_TYPE")
if platformType == "" {
    platformType = platform.PlatformTypeOpenDataHub // default
}

p, err := factory.New(platformType, cli)
if err != nil {
    return fmt.Errorf("failed to create platform: %w", err)
}

// Initialize platform
if err := p.Init(ctx); err != nil {
    return fmt.Errorf("platform initialization failed: %w", err)
}

// Get platform display name (implements fmt.Stringer)
log.Info("Platform initialized", "name", p.String())
// Output: "Open Data Hub", "OpenShift AI Self-Managed", etc.
```

### Using Platform Validators

```go
// Get platform-specific validators
validator := p.Validator()

// Register DSCI validator with webhook server
dsciHandler := validator.DSCInitializationValidator()
webhookServer.Register("/validate-dsci", dsciHandler)

// Register DSC validator with webhook server
dscHandler := validator.DataScienceClusterValidator()
webhookServer.Register("/validate-dsc", dscHandler)
```

## Environment Variables

- `ODH_PLATFORM_TYPE` - Specifies which platform variant to use
  - Valid values: `selfmanaged`, `opendatahub`, `managed`, `vanilla`
  - Default: `opendatahub`

## Extension Points

All platform implementations currently contain placeholder logic marked with `// TODO` comments. These are extension points for adding platform-specific behavior:

1. **Upgrade Logic** - Custom upgrade procedures per platform
2. **Initialization** - Platform-specific setup during operator startup
3. **Runtime Logic** - Platform-specific operations during reconciliation
4. **Validation Rules** - Custom CR validation per platform

## Adding a New Platform Variant

1. Create a new package under `pkg/platform/<variant>/`
2. Implement the `Platform` interface in `<variant>.go`
3. Implement the `Validator` interface in `validator.go`
4. Add the new variant to `pkg/platform/factory/factory.go`
5. Add the platform constant to `pkg/platform/platform.go`

## Design Decisions

### Factory Pattern
The factory package is separate from the platform package to avoid circular imports. Platform implementations import the `platform` package for interfaces, while the factory imports all implementations.

### Validator Naming
Validation handlers are named `dsciHandler` and `dscHandler` to clearly indicate their purpose (DSCInitialization and DataScienceCluster validation respectively).

### Interface-First Design
All functionality is defined through interfaces, allowing easy mocking for testing and enabling future platform variants without modifying core operator code.