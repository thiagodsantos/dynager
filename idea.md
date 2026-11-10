# DynamoDB Web Interface

## Create plan to create a modern dynamodb web interface

Pages:
- Setup/Configuration
 - Setup should have test connection
- List connections
- List tables and own records
 - Can use filter
 - Can edit/delete record

### Tech Stack:
- Frontend
 - Use PNPM, TailwindCSS, Vite, React with Atomic Design and Typescript
 - Use UI from https://github.com/TailAdmin/tailadmin-free-tailwind-dashboard-template
 - Light and Dark themes
 - Use path aliases in import, avoid relative path
 - Use latest Node.js version
 - Use routing pages
- Backend
 - Use Golang with SOLID, DRY and YAGNI principles
 - Use REST with router: https://github.com/go-chi/chi
   - Should have versioning control, using header version. Ex: x-api-version: v1
 - Unit/Integration tests
   - Docker compose to run integration tests

 ### AI
 - Create agents with skills from https://github.com/mattpocock/skills/tree/main

Should be a monorepo (frontend and backend)
Makefile to all commands (install, run, build, etc...)
Create README.md (What is, Tech Stack, Architecture, How to use, etc...)
Create icon/image to application
Create file to MIT License
