## Запуск
1. В корневой папке, где находится `docker-compose.yml`, необходимо запустить все контейнеры: `docker-compose up -d`.
2. Подключиться к БД PostgreSQL, развёрнутой в Docker (порт `5432`).
3. Создать БД, применив инструкции из `database-initialization/1-create_db.sql`.
4. Создать таблицу, применив инструкции из `database-initialization/2-initialize_db.sql`.
5. Использовать Postman-коллекцию `SocialNetwork hw01.postman_collection.json`.