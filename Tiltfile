# Tiltfile for Billy Wu development environment

# Load environment variables from client/.env
load('ext://dotenv', 'dotenv')
load('ext://restart_process', 'docker_build_with_restart')

dotenv('./.env')

# Configuration - use environment variables with defaults
SERVER_PORT = os.getenv('SERVER_PORT', '8280')
CLIENT_PORT = os.getenv('CLIENT_PORT', '3010')
VALKEY_PORT = os.getenv('VALKEY_PORT', '6379')

# Development mode toggle
DEV_MODE = True

# Go Server with Air hot reloading - Volume mount approach
docker_build(
    'billy-wu-server-dev',
    context='./server',
    dockerfile='./server/Dockerfile.dev',
    target='development',
    # No live_update - use volume mounts instead
    ignore=[
        'tmp/', 
        '*.log', 
        'main',
        '.git/',
        'Dockerfile*',
        '.dockerignore',
    ]
)

# SolidJS Client with Vite hot reloading  
docker_build(
    'billy-wu-client-dev',
    context='./client',
    dockerfile='./client/Dockerfile.dev',
    live_update=[
        # Sync only source directories
        sync('./client/src', '/app/src'),
        sync('./client/public', '/app/public'),
        # Sync config files individually
        sync('./client/package.json', '/app/package.json'),
        sync('./client/package-lock.json', '/app/package-lock.json'),
        sync('./client/vite.config.ts', '/app/vite.config.ts'),
        sync('./client/tsconfig.json', '/app/tsconfig.json'),
        sync('./client/index.html', '/app/index.html'),
        sync('./client/.env', '/app/.env'),
        # Run npm install when package.json changes
        run('npm install', trigger=['./client/package.json', './client/package-lock.json']),
    ],
    ignore=[
        'node_modules/', 
        'dist/', 
        'build/', 
        '.vite/',
        '.*.swp',
        '.*.swo',
        '*~',
        '.DS_Store',
        '.git/',
        '.gitignore',
        'Dockerfile*',
        '.dockerignore',
    ]
)

# Valkey database service
docker_build(
    'billy-wu-valkey-dev',
    context='./database/valkey',
    dockerfile='./database/valkey/Dockerfile.dev',
    live_update=[
        # Sync configuration changes
        sync('./database/valkey/valkey.conf', '/usr/local/etc/valkey/valkey.conf'),
        # Restart container when config changes (Valkey needs restart for config changes)
        restart_container(),
    ],
    ignore=[
        '.*.swp',
        '.*.swo',
        '*~',
        '.DS_Store',
        '.git/',
        '.gitignore',
    ]
)

# Use docker-compose for orchestration
docker_compose('./docker-compose.dev.yml')

# Configure resources with better organization
dc_resource('valkey',
    labels=['database'],
    resource_deps=[],
)

dc_resource('server', 
    labels=['backend'],
    resource_deps=['valkey'],
)

dc_resource('client', 
    labels=['frontend'],
    resource_deps=['server']
)

# Development utilities
if DEV_MODE:
    # Add a local resource for running server tests
    local_resource(
        'server-tests',
        cmd='cd server && go test ./...',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['tests'],
        auto_init=False,  # Don't run automatically
        trigger_mode=TRIGGER_MODE_MANUAL  # Run manually via Tilt UI
    )

    # Server test coverage
    local_resource(
        'server-tests-coverage',
        cmd='cd server && go test -cover ./...',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['tests'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # HTML coverage report with auto-open
    local_resource(
        'server-coverage-html',
        cmd='cd server && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['tests'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )
    
    # Add linting
    local_resource(
        'server-lint',
        cmd='cd server && golangci-lint run',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['linting'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    local_resource(
        'client-lint',
        cmd='cd client && npm run lint:check',
        deps=['./client/src'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['linting'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )
    
    # Client tests
    local_resource(
        'client-tests',
        cmd='cd client && npm run test',
        deps=['./client/src'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['tests'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Valkey utilities
    local_resource(
        'valkey-info',
        cmd='docker compose -f docker-compose.dev.yml exec valkey valkey-cli info',
        labels=['database'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Database migration commands
    local_resource(
        'migrate-up',
        cmd='docker compose -f docker-compose.dev.yml exec server go run cmd/migration/main.go up',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['database'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server'] 
    )

    local_resource(
        'migrate-down',
        cmd='docker compose -f docker-compose.dev.yml exec server go run cmd/migration/main.go down 1',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['database'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server']
    )

    local_resource(
        'migrate-seed',
        cmd='docker compose -f docker-compose.dev.yml exec server go run cmd/migration/main.go seed',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['database'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server']
    )


print("🚀 Billy Wu Development Environment")
print("📊 Tilt Dashboard: http://localhost:10350" )
print("🔧 Server API: http://localhost:%s" % SERVER_PORT)
print("🎨 Client App: http://localhost:%s" % CLIENT_PORT)
print("💾 Valkey DB: localhost:%s" % VALKEY_PORT)
print("💡 Hot reloading enabled for all services!")
print("🧪 Manual test/lint resources available in Tilt UI")

# Development shortcuts
print("\n📋 Quick Commands:")
print("• tilt trigger server-tests    - Run Go tests")
print("• tilt trigger server-lint     - Run Go linting") 
print("• tilt trigger client-tests    - Run frontend tests")
print("• tilt trigger migrate-up      - Run database migrations")
print("• tilt trigger migrate-down    - Rollback 1 migration")
print("• tilt trigger migrate-seed    - Reset and seed database")
print("• tilt trigger valkey-cli      - Access Valkey CLI")
print("• tilt trigger valkey-info     - Show Valkey info")
print("• tilt down                    - Stop all services")
print("• tilt up --stream             - Start with streaming logs")
