# Guide: เพิ่ม Service ใหม่

คู่มือสำหรับเพิ่ม service/feature ใหม่เข้าสู่ project ตามโครงสร้าง Clean Architecture

---

## สารบัญ

1. [Project Structure](#project-structure)
2. [Dependency Flow](#dependency-flow)
3. [ขั้นตอนการเพิ่ม Service ใหม่](#ขั้นตอนการเพิ่ม-service-ใหม่)
4. [ตัวอย่าง: เพิ่ม Course Service](#ตัวอย่าง-เพิ่ม-course-service)
5. [Response Helper](#response-helper)
6. [Checklist](#checklist)

---

## Project Structure

```
.
├── cmd/api/main.go                          # Entry point
├── internal/
│   ├── config/                              # Configuration (env vars)
│   ├── domain/
│   │   ├── entity/                          # Business entities (ไม่มี json tag)
│   │   └── usecase/                         # Business logic + interfaces
│   └── delivery/http/
│       ├── adapter/                         # Framework adapters (Fiber → Responder)
│       ├── dto/                             # Data Transfer Objects (json tag อยู่ที่นี่)
│       ├── handler/                         # HTTP handlers
│       ├── middleware/                       # HTTP middlewares
│       ├── router/                          # Route definitions
│       └── server/                          # Server bootstrap & DI wiring
├── pkg/
│   ├── port/                                # Interfaces (Responder, etc.)
│   └── response/                            # Response helpers (framework-agnostic)
```

---

## Dependency Flow

```
Handler → Usecase (interface) → Entity
   ↓           ↓
  DTO     Repository (interface) → DB/External
   ↓
Response Helper → port.Responder (interface) → Adapter (Fiber)
```

**กฎสำคัญ:**
- `domain/` ห้าม import `delivery/` หรือ framework ใดๆ
- `entity` ไม่มี `json` tags — เป็น pure Go struct
- `dto` อยู่ใน delivery layer — มี `json` tags + mapper functions
- `usecase` define interface ของตัวเอง + repository interface (ถ้ามี)
- `handler` เรียก usecase ผ่าน interface เท่านั้น

---

## ขั้นตอนการเพิ่ม Service ใหม่

### Step 1: สร้าง Entity

สร้างไฟล์ใน `internal/domain/entity/`

```go
// internal/domain/entity/{name}.go
package entity

type {Name} struct {
    ID   string
    Name string
    // ... fields ตาม business (ไม่มี json tag)
}
```

### Step 2: สร้าง Repository Interface (ถ้าต้องติดต่อ DB)

สร้างไฟล์ใน `internal/domain/usecase/` หรือแยก `internal/domain/repository/`

```go
// internal/domain/usecase/{name}_repository.go
package usecase

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

type {Name}Repository interface {
    FindAll() ([]entity.{Name}, error)
    FindByID(id string) (*entity.{Name}, error)
    Create(e *entity.{Name}) error
    Update(e *entity.{Name}) error
    Delete(id string) error
}
```

### Step 3: สร้าง Usecase

สร้างไฟล์ใน `internal/domain/usecase/`

```go
// internal/domain/usecase/{name}_usecase.go
package usecase

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// {Name}Usecase defines business logic interface.
type {Name}Usecase interface {
    GetAll() ([]entity.{Name}, error)
    GetByID(id string) (*entity.{Name}, error)
    Create(e *entity.{Name}) error
}

type {name}Usecase struct {
    repo {Name}Repository
}

// New{Name}Usecase creates a new usecase instance.
func New{Name}Usecase(repo {Name}Repository) {Name}Usecase {
    return &{name}Usecase{repo: repo}
}

func (u *{name}Usecase) GetAll() ([]entity.{Name}, error) {
    return u.repo.FindAll()
}

func (u *{name}Usecase) GetByID(id string) (*entity.{Name}, error) {
    return u.repo.FindByID(id)
}

func (u *{name}Usecase) Create(e *entity.{Name}) error {
    // business validation here
    return u.repo.Create(e)
}
```

### Step 4: สร้าง DTO + Mapper

สร้างไฟล์ใน `internal/delivery/http/dto/`

```go
// internal/delivery/http/dto/{name}_dto.go
package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// --- Response DTO ---

type {Name}Response struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func To{Name}Response(e entity.{Name}) {Name}Response {
    return {Name}Response{
        ID:   e.ID,
        Name: e.Name,
    }
}

func To{Name}ListResponse(list []entity.{Name}) []{Name}Response {
    result := make([]{Name}Response, len(list))
    for i, e := range list {
        result[i] = To{Name}Response(e)
    }
    return result
}

// --- Request DTO ---

type Create{Name}Request struct {
    Name string `json:"name" validate:"required"`
}

func (r *Create{Name}Request) ToEntity() entity.{Name} {
    return entity.{Name}{
        Name: r.Name,
    }
}
```

### Step 5: สร้าง Handler

สร้างไฟล์ใน `internal/delivery/http/handler/`

```go
// internal/delivery/http/handler/{name}_handler.go
package handler

import (
    "github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
    "github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
    "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
    "github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
    "github.com/gofiber/fiber/v2"
)

type {Name}Handler struct {
    usecase usecase.{Name}Usecase
}

func New{Name}Handler(uc usecase.{Name}Usecase) *{Name}Handler {
    return &{Name}Handler{usecase: uc}
}

func (h *{Name}Handler) GetAll(c *fiber.Ctx) error {
    r := adapter.NewFiberResponder(c)

    list, err := h.usecase.GetAll()
    if err != nil {
        return response.InternalError(r, err.Error())
    }

    return response.OK(r, dto.To{Name}ListResponse(list))
}

func (h *{Name}Handler) GetByID(c *fiber.Ctx) error {
    r := adapter.NewFiberResponder(c)
    id := c.Params("id")

    item, err := h.usecase.GetByID(id)
    if err != nil {
        return response.NotFound(r, "not found")
    }

    return response.OK(r, dto.To{Name}Response(*item))
}

func (h *{Name}Handler) Create(c *fiber.Ctx) error {
    r := adapter.NewFiberResponder(c)

    var req dto.Create{Name}Request
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(r, "invalid request body")
    }

    e := req.ToEntity()
    if err := h.usecase.Create(&e); err != nil {
        return response.InternalError(r, err.Error())
    }

    return response.Created(r, dto.To{Name}Response(e))
}
```

### Step 6: ลงทะเบียน Route

แก้ไฟล์ `internal/delivery/http/router/router.go`

```go
func SetupRoutes(
    app *fiber.App,
    healthHandler *handler.HealthHandler,
    versionHandler *handler.VersionHandler,
    {name}Handler *handler.{Name}Handler,   // ← เพิ่ม parameter
) {
    api := app.Group("/api/v1")

    // existing
    api.Get("/status", healthHandler.GetStatus)
    api.Get("/version", versionHandler.GetVersion)

    // ← เพิ่ม routes ใหม่
    {name}s := api.Group("/{name}s")
    {name}s.Get("/", {name}Handler.GetAll)
    {name}s.Get("/:id", {name}Handler.GetByID)
    {name}s.Post("/", {name}Handler.Create)
}
```

### Step 7: Wire Dependencies ใน Server

แก้ไฟล์ `internal/delivery/http/server/server.go`

```go
func Start(cfg *config.Config) {
    app := fiber.New(fiber.Config{AppName: cfg.AppName})
    middleware.SetupMiddlewares(app)

    // usecases
    healthUC  := usecase.NewHealthUsecase()
    versionUC := usecase.NewVersionUsecase(cfg.AppName, cfg.AppVersion, cfg.AppEnv)
    {name}UC  := usecase.New{Name}Usecase({name}Repo)  // ← เพิ่ม

    // handlers
    healthHandler  := handler.NewHealthHandler(healthUC)
    versionHandler := handler.NewVersionHandler(versionUC)
    {name}Handler  := handler.New{Name}Handler({name}UC) // ← เพิ่ม

    // routes
    router.SetupRoutes(app, healthHandler, versionHandler, {name}Handler) // ← เพิ่ม parameter

    addr := fmt.Sprintf(":%s", cfg.Port)
    log.Printf("Server starting on %s (env=%s)", addr, cfg.AppEnv)
    app.Listen(addr)
}
```

---

## ตัวอย่าง: เพิ่ม Course Service

สมมติต้องเพิ่ม API จัดการวิชาเรียน — ไฟล์ที่ต้องสร้าง/แก้ไข:

| # | ไฟล์ | สร้าง/แก้ |
|---|------|-----------|
| 1 | `internal/domain/entity/course.go` | สร้างใหม่ |
| 2 | `internal/domain/usecase/course_repository.go` | สร้างใหม่ |
| 3 | `internal/domain/usecase/course_usecase.go` | สร้างใหม่ |
| 4 | `internal/delivery/http/dto/course_dto.go` | สร้างใหม่ |
| 5 | `internal/delivery/http/handler/course_handler.go` | สร้างใหม่ |
| 6 | `internal/delivery/http/router/router.go` | แก้ไข |
| 7 | `internal/delivery/http/server/server.go` | แก้ไข |

---

## Response Helper

ใช้ `pkg/response` สำหรับส่ง response ทุกครั้ง — ห้ามเขียน `c.JSON(...)` ตรง ๆ ใน handler

```go
import (
    "github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
    "github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
)

// ภายใน handler function:
r := adapter.NewFiberResponder(c)

// Success
response.OK(r, data)              // 200
response.Created(r, data)         // 201
response.NoContent(r)             // 204

// Error
response.BadRequest(r, "msg")          // 400
response.Unauthorized(r, "msg")        // 401
response.Forbidden(r, "msg")           // 403
response.NotFound(r, "msg")            // 404
response.Conflict(r, "msg")            // 409
response.UnprocessableEntity(r, "msg") // 422
response.InternalError(r, "msg")       // 500
response.ValidationError(r, errors)    // 422 + field errors
```

**Response format:**

```json
{
  "success": true,
  "data": { ... }
}
```

```json
{
  "success": false,
  "error": {
    "code": 400,
    "message": "invalid request body"
  }
}
```

---

## Checklist

เมื่อเพิ่ม service ใหม่ ตรวจสอบตาม checklist นี้:

- [ ] Entity ไม่มี `json` tag และไม่ import package นอก domain
- [ ] Usecase define interface + private struct implement
- [ ] Repository interface อยู่ใน domain layer (ถ้ามี)
- [ ] DTO มี `json` tag + mapper function (`To{Name}Response`, `ToEntity`)
- [ ] Handler ใช้ `adapter.NewFiberResponder(c)` + `response.*` helper
- [ ] Handler เรียก usecase ผ่าน interface เท่านั้น
- [ ] Route ลงทะเบียนใน `router.go`
- [ ] Dependency wiring ใน `server.go`
- [ ] `go build ./...` ผ่าน
- [ ] ทดสอบ endpoint ด้วย `curl`
