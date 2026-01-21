# FPart

## Overview

FPart is a project management system focused on developers and small teams.
The product is positioned as a Russian-language alternative to Jira, with an emphasis on simplicity, clear domain boundaries, and predictable backend behavior.

At the current stage, the project is developed as an MVP with a limited but well-defined scope.

---

## Target Audience

* Developers
* Small engineering teams
* Early-stage internal projects

---

## MVP Scope

### Supported functionality

* Project management

  * Create projects
  * Delete projects
  * Add and manage project participants

* User information

  * Get current user profile
  * Get project details
  * Get list of user projects

* Authentication

  * Login via Google OAuth

### Out of scope (MVP)

* Authentication methods other than Google OAuth
* Advanced task management
* Project configuration and permissions
* Analytics and metrics collection

---

## Architecture

* Backend follows a layered architecture with clear separation between:

  * Transport layer
  * Application logic
  * Domain logic
  * Infrastructure

* Monolithic application with explicit internal boundaries

* API-oriented backend, frontend communicates exclusively via HTTP

More details can be found in `/docs/architecture.md`.

---

## Technology Stack

**Frontend**

* Vue 3
* Vite
* TypeScript

**Backend**

* Go
* fasthttp

**Storage**

* PostgreSQL
* Redis

**Infrastructure**

* Nginx
* Docker

**Environments**

* Development
* Production

---

## Local Development

```bash
make run
```

---

## Tests

```bash
make test
```

---

## Project Status

The project is under active development.
The current focus is on stabilizing core flows, error handling, and API contracts.
