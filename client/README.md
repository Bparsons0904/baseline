# Billy Wu Client

A modern, responsive SolidJS frontend application with TypeScript, featuring a comprehensive design system, authentication, and real-time WebSocket communication.

## 🏗️ Technology Stack

- **Framework**: [SolidJS](https://www.solidjs.com/) with TypeScript
- **Build Tool**: [Vite](https://vitejs.dev/) for fast development and building
- **Styling**: SCSS with CSS Modules and design tokens
- **Routing**: [@solidjs/router](https://github.com/solidjs/solid-router)
- **State Management**: [Solid Query](https://tanstack.com/query/latest/docs/framework/solid/overview) + Context API
- **HTTP Client**: Axios with custom API service layer
- **WebSockets**: Custom WebSocket context for real-time features
- **Linting**: ESLint with TypeScript support

## 📁 Project Structure

```
client/
├── src/
│   ├── components/              # Reusable UI components
│   │   ├── common/              # Generic components
│   │   │   ├── Button/          # Button component with variants
│   │   │   └── forms/           # Form components (TextInput, etc.)
│   │   └── layout/              # Layout components
│   │       └── Navbar/          # Navigation bar
│   ├── context/                 # React-style context providers
│   │   ├── AuthContext.tsx      # Authentication state management
│   │   └── WebSocketContext.tsx # WebSocket connection management
│   ├── pages/                   # Page components
│   │   ├── Auth/                # Authentication pages
│   │   └── Home/                # Home page
│   ├── services/                # API and external service integrations
│   │   ├── api/                 # API service layer
│   │   └── env.service.ts       # Environment configuration
│   ├── styles/                  # Global styles and design system
│   │   ├── _colors.scss         # Color system and tokens
│   │   ├── _variables.scss      # Design tokens (spacing, typography, etc.)
│   │   ├── _mixins.scss         # SCSS mixins and utilities
│   │   ├── _reset.scss          # CSS reset
│   │   └── global.scss          # Global styles import
│   ├── types/                   # TypeScript type definitions
│   ├── App.tsx                  # Root application component
│   ├── Routes.tsx               # Application routing
│   └── index.tsx                # Application entry point
├── public/                      # Static assets
├── index.html                   # HTML template
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── vite.config.ts               # Vite build configuration
├── eslint.config.js             # ESLint configuration
├── StyleGuide.md                # Comprehensive style guide
├── .env                         # Environment variables
├── .dockerignore                # Docker ignore patterns
└── Dockerfile.dev               # Development Docker image
```

## 🚀 Getting Started

### Prerequisites

- Node.js v22 (specified in `.nvmrc`)
- npm (comes with Node.js)

### Local Development

1. **Install Dependencies**:

   ```bash
   npm install
   ```

2. **Start Development Server**:

   ```bash
   npm run dev
   ```

3. **Access Application**:
   - Application: http://localhost:3010
   - Hot reloading enabled automatically

### Docker Development

The recommended way is through the main project's Tilt setup, but you can also run the client container directly:

```bash
# Build and run development container
docker build -f Dockerfile.dev -t billy-wu-client-dev .
docker run -p 3010:3010 billy-wu-client-dev
```

## 🎨 Design System & Styling

### Design Philosophy

The client follows a **mobile-first**, **component-based** design approach with a comprehensive design token system. See [StyleGuide.md](./StyleGuide.md) for detailed guidelines.

### Design Tokens

#### Color System

```scss
// Primary colors
$color-primary-500: #6366f1; // Main brand color
$color-primary-600: #4f46e5; // Darker variant

// Semantic colors
$text-default: $color-gray-900; // Default text
$bg-primary: $color-primary-600; // Primary backgrounds
$border-focus: $color-primary-500; // Focus states
```

#### Spacing Scale

```scss
$spacing-xs: 0.25rem; // 4px
$spacing-sm: 0.5rem; // 8px
$spacing-md: 1rem; // 16px
$spacing-lg: 1.5rem; // 24px
$spacing-xl: 2rem; // 32px
```

#### Typography Scale

```scss
$font-size-sm: 0.875rem; // 14px
$font-size-md: 1rem; // 16px
$font-size-lg: 1.125rem; // 18px
$font-size-xl: 1.25rem; // 20px
```

### CSS Architecture

#### CSS Modules

All component styles use CSS Modules for encapsulation:

```tsx
import styles from "./Button.module.scss";

export const Button = () => <button className={styles.button}>Click me</button>;
```

#### Responsive Design

Mobile-first approach with breakpoint mixins:

```scss
.component {
  // Mobile styles (default)
  padding: $spacing-sm;

  // Tablet and up
  @include breakpoint(md) {
    padding: $spacing-md;
  }

  // Desktop and up
  @include breakpoint(lg) {
    padding: $spacing-lg;
  }
}
```

#### Design Token Usage

```scss
// Import design tokens
@use "../../styles/variables" as *;
@use "../../styles/mixins" as *;
@use "../../styles/colors" as *;

.component {
  padding: $spacing-md;
  color: $text-default;
  background: $bg-surface;
  border-radius: $border-radius-md;
}
```

## 🧩 Component System

### Common Components

#### Button Component

```tsx
<Button variant="primary" size="md" onClick={handleClick}>
  Submit
</Button>
```

**Variants**: `primary`, `secondary`, `tertiary`, `danger`  
**Sizes**: `sm`, `md`, `lg`

#### TextInput Component

```tsx
<TextInput
  label="Email"
  type="email"
  autoComplete="email"
  onBlur={(value) => handleUpdate("email", value)}
/>
```

### Layout Components

#### NavBar

- Responsive navigation with mobile-first design
- Authentication state integration
- Active link highlighting

## 🔗 Routing & Navigation

Routes are defined in `Routes.tsx` using `@solidjs/router`:

```tsx
export const Routes = () => (
  <>
    <Route path="/" component={HomePage} />
    <Route path="/login" component={LoginPage} />
  </>
);
```

**Features**:

- Lazy-loaded page components
- Type-safe routing
- Authentication guards (ready for implementation)

## 🔐 Authentication

### AuthContext

Centralized authentication state management:

```tsx
const { isAuthenticated, user, login, logout } = useAuth();

// Login
await login({ login: "username", password: "password" });

// Logout
logout();

// Check authentication status
if (isAuthenticated()) {
  // User is logged in
}
```

**Features**:

- JWT token management via HTTP-only cookies
- Automatic session validation
- Protected route support
- User state persistence

## 🌐 API Integration

### API Service Layer

Centralized API communication with axios **and automatic token capture**:

```tsx
// API calls automatically capture JWT tokens via response interceptors
// GET request
const user = await getApi<User>("users");

// POST request
const result = await postApi<User, LoginCredentials>(
  "users/login",
  credentials,
);
// X-Auth-Token header automatically captured for WebSocket authentication
```

### Solid Query Integration

Reactive data fetching with caching:

```tsx
const userQuery = useQuery(() => ({
  queryKey: ["user"],
  queryFn: () => getApi<{ user: User }>("users"),
}));
```

**Features**:

- Automatic caching and invalidation
- Loading and error states
- Optimistic updates
- Background refetching

## 🔄 Real-time Features

### WebSocket Integration

Custom WebSocket context for real-time communication with **automatic JWT authentication**:

````tsx
const { isConnected, sendMessage, lastMessage } = useWebSocket();

// Connection automatically uses JWT token from auth context
// No manual token management required

// Send message
sendMessage("Hello server!");

// Connection status
if (isConnected()) {
  // WebSocket is authenticated and connected
}

**Features**:

- Automatic reconnection
- Connection state management
- Message queuing
- Error handling

## 🛠️ Development Tools

### Available Scripts

```bash
# Development
npm run dev          # Start development server
npm run build        # Build for production
npm run serve        # Preview production build

# Code Quality
npm run lint         # Run ESLint
npm run lint:check   # Check linting without fixing
````

### Path Aliases

Configured TypeScript path aliases for clean imports:

```tsx
import { Button } from "@components/common/Button/Button";
import { useAuth } from "@context/AuthContext";
import { colors } from "@styles/colors";
```

**Available Aliases**:

- `@styles/*` → `src/styles/*`
- `@components/*` → `src/components/*`
- `@layout/*` → `src/components/layout/*`
- `@pages/*` → `src/pages/*`
- `@hooks/*` → `src/hooks/*`
- `@services/*` → `src/services/*`
- `@context/*` → `src/context/*`

### Hot Module Replacement

Vite provides instant hot reloading for:

- Component changes
- Style updates
- TypeScript changes
- Environment variable updates

## 🧪 Testing & Quality

### Linting Configuration

ESLint is configured for TypeScript and SolidJS:

```javascript
// eslint.config.js
export default defineConfig([
  // TypeScript support
  ...tseslint.configs.recommended,
  // CSS and JSON linting
  { files: ["**/*.css"], extends: ["css/recommended"] },
  { files: ["**/*.json"], extends: ["json/recommended"] },
]);
```

### Code Quality Standards

- TypeScript strict mode enabled
- ESLint rules for consistent code style
- CSS/SCSS linting for style consistency
- Import organization and unused import detection

## ⚙️ Configuration

### Environment Variables

Configured in `.env`:

```bash
VITE_API_URL=http://localhost:8280
VITE_WS_URL=ws://localhost:8280/ws
VITE_ENV=local
```

### Vite Configuration

Optimized for SolidJS development:

```typescript
export default defineConfig({
  plugins: [solidPlugin()],
  server: {
    port: 3010,
    host: "0.0.0.0",
    hmr: { port: 3010 },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `
          @use "@styles/variables" as *;
          @use "@styles/mixins" as *;
          @use "@styles/colors" as *;
        `,
      },
    },
  },
});
```

## 🏗️ Build & Deployment

### Development Build

```bash
npm run build
```

### Production Optimization

- Tree shaking for smaller bundles
- CSS optimization and minification
- Asset optimization
- TypeScript compilation

### Docker Production

```dockerfile
FROM node:22-alpine
# Multi-stage build for optimized production image
```

## 🔧 Troubleshooting

### Common Issues

1. **Port 3010 Already in Use**:

   ```bash
   # Find and kill process
   lsof -i :3010
   kill -9 <PID>
   ```

2. **Node Version Mismatch**:

   ```bash
   # Use correct Node version
   nvm use 22
   # Or install if not available
   nvm install 22
   ```

3. **SCSS Import Errors**:

   - Ensure design tokens are imported in component SCSS files
   - Check path aliases in `vite.config.ts`

4. **TypeScript Errors**:

   - Run `npm run build` to see all TypeScript errors
   - Check path aliases in `tsconfig.json`

5. **WebSocket Connection Issues**:
   - Verify server is running on port 8280
   - Check WebSocket URL in environment variables

### Development Tips

- Use browser dev tools for SolidJS debugging
- Check Vite dev server logs for build issues
- Use TypeScript strict mode for better code quality
- Monitor network tab for API request debugging

## 🤝 Contributing

1. **Component Development**:

   - Follow the design system guidelines in `StyleGuide.md`
   - Use CSS Modules for component styling
   - Implement responsive design with mobile-first approach

2. **Code Standards**:

   - Run ESLint before committing
   - Use TypeScript strict mode
   - Follow established file organization patterns

3. **Style Guidelines**:
   - Use design tokens for consistency
   - Follow BEM-like naming in CSS Modules
   - Implement accessible components

## 📚 Additional Resources

- [SolidJS Documentation](https://www.solidjs.com/docs)
- [Vite Documentation](https://vitejs.dev/guide/)
- [TanStack Query for Solid](https://tanstack.com/query/latest/docs/framework/solid/overview)
- [Design System Style Guide](./StyleGuide.md)
