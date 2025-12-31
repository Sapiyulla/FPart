```
Frontend/
├── Dockerfile
├── package.json
├── package-lock.json / pnpm-lock.yaml
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
├── index.html
├── .eslintrc.cjs
├── .env
├── .env.development
├── .env.production
├── node_modules/
└── src/
    ├── main.tsx
    ├── vite-env.d.ts
    │
    ├── app/
    │   ├── App.tsx
    │   ├── router.tsx
    │   ├── providers/
    │   │   ├── QueryProvider.tsx
    │   │   ├── ThemeProvider.tsx
    │   │   └── index.ts
    │   └── styles/
    │       ├── globals.css
    │       ├── reset.css
    │       └── variables.css
    │
    ├── pages/
    │   ├── Login/
    │   │   ├── LoginPage.tsx
    │   │   ├── LoginPage.module.css
    │   │   └── index.ts
    │   └── Dashboard/
    │       ├── DashboardPage.tsx
    │       └── index.ts
    │
    ├── features/
    │   ├── auth/
    │   │   ├── api.ts
    │   │   ├── model.ts
    │   │   ├── hooks.ts
    │   │   └── index.ts
    │   └── user/
    │       ├── api.ts
    │       ├── model.ts
    │       └── index.ts
    │
    ├── entities/
    │   └── user/
    │       ├── types.ts
    │       └── index.ts
    │
    ├── shared/
    │   ├── ui/
    │   │   ├── Button/
    │   │   │   ├── Button.tsx
    │   │   │   ├── Button.module.css
    │   │   │   └── index.ts
    │   │   └── Input/
    │   │       ├── Input.tsx
    │   │       ├── Input.module.css
    │   │       └── index.ts
    │   │
    │   ├── lib/
    │   │   ├── http.ts
    │   │   ├── env.ts
    │   │   └── constants.ts
    │   │
    │   └── hooks/
    │       ├── useDebounce.ts
    │       └── useThrottle.ts
    │
    └── assets/
        ├── images/
        └── icons/
```

**Назначение каталогов**

```
Frontend/                dev-контейнер и конфигурация
src/                     исходный код приложения
app/                     инициализация, провайдеры, глобальные стили
pages/                   маршрутизируемые страницы
features/                бизнес-кейсы и пользовательские сценарии
entities/                доменные сущности и типы
shared/ui/               переиспользуемые UI-компоненты
shared/lib/              утилиты, клиенты, конфигурации
shared/hooks/            общие React-хуки
assets/                  статические ресурсы
```

**Правила архитектуры**

```
pages        не содержат бизнес-логики
features     не знают о страницах
entities     не используют React
shared/ui    не содержит бизнес-логики
app          не содержит feature-кода
```

**Импортная иерархия**

```
shared → entities → features → pages → app
```

**Принцип разработки**

```
одна feature = один пользовательский сценарий
один компонент = одна ответственность
отсутствие циклических импортов
```

**Рабочий цикл**

```
правка файла → volume → vite HMR → браузер
```
