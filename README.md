<h1> Система для работы с постами и комментариями </h1>

В этом проекте реализована система для работы с постами и комментариями аналогичная блогам/формам

Вдохновлено примером из библиотеки https://github.com/99designs/gqlgen/tree/master/_examples/mini-habr-with-subscriptions

В проекте реализовано:
- хранение данных в PostgreSQL с использованием sqlc для генерации типизированного кода;
- поддержка in‑memory хранилища для быстрого прототипирования и тестов;
- древовидные комментарии с помощью path (массив идентификаторов) для эффективной сортировки и выборки;
- подписка на новые комментарии через GraphQL Subscriptions (WebSocket);
- Docker Compose для быстрого развертывания
  
Стек технологий

    Go 1.26.1

    GraphQL – gqlgen

    PostgreSQL 17

    pgx – драйвер для работы с PostgreSQL

    sqlc – генерация кода для работы с БД

    goose – миграции

    Docker / Docker Compose

    testify – для тестирования

Как запустить у себя:
1. Клонируйте репозиторий
   ```
   git clone https://github.com/ganzhinza/postsys.git
   ```
2. Насторить переменные окружения(через файл .env)
3. Запуск через Docker Compose
   ```
   docker-compose up
   ```
   После запуска GraphQL Playground будет доступен по адресу: http://localhost:8080/
   P.S. Для продакшена лучше отключать, но удобно для тестирования

Примеры запросов:
- Получить все посты
  ```
  query {
    posts {
      id
      title
      content
      allowComments
      createdAt
    }
  }
  ```
- Получить пост с комментариями и пагинацией:
  ```
  query {
    post(id: 1) {
      id
      title
      comments(limit: 5, offset: 0) {
        comments {
          id
          content
          branch(limit: 10) {
            comments { id content }
            hasNext
          }
        }
        hasNext
        totalCount
      }
    }
  }
  ```
- Создать пост
  ```
  mutation {
    createPost(input: {
      authorID: 1
      title: "Hello"
      content: "World"
      allowComments: true
    }) {
      id
      title
    }
  }
  ```
- Создать комментарий
  ```
  mutation {
    createComment(input: {
      postID: 1
      content: "Great post!"
      authorID: 2
      parentID: null   
    }) {
      id
      content
      branch {
        comments { id }
      }
    }
  }
  ```
 - Включить/отключить комментарии у поста (может менять только автор):
   ```
   mutation {
      updatePostCommentsAvailability(postID: 1, userID: 1, availability: false) {
        id
        allowComments
      }
    }
   ```
- Подписаться на комментарии к посту:
  ```
  subscription {
    addComment(postID: 1) {
      id
      content
      authorID
      createdAt
    }
  }
  ```
