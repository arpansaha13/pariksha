# GitHub Copilot Instructions

## General Coding Guidelines
- Write clear and concise code, following best practices for Vue.js, Nuxt 3, and Go.
- Ensure function names, variable names, and comments improve code readability.
- Prefer explicit over implicit logic to enhance maintainability.

## Function Documentation
- When creating a new function, add a short comment before it describing its main purpose.

Example:

```go
// Add returns the sum of two integers.
func Add(a, b int) int {
    return a + b
}
```

## Project coding standards

### TypeScript Guidelines
- Use TypeScript for all new code
- Follow functional programming principles where possible
- Use interfaces for data structures and type definitions
- Prefer immutable data (const, readonly)
- Use optional chaining (?.) and nullish coalescing (??) operators

## Security Best Practices

- Avoid hardcoding sensitive information.
- Sanitize user input and prevent injection attacks.
- Use environment variables for configuration.
