# URL-Shortener service written in golang with github actions pipeline
Сервис позволяет сохранять и получать ссылку используя, придумааный вами alias (если не передать в запросе ваш alias он сгенерируется автоматически и вернется вам в ответе).
## Для запуска пайплайна:
1) Создать необходимые secrets в github action:  

   **Credentials от вашего dockerhub:**

    * DOCKER_USERNAME
    * DOCKER_PASSWORD   
    
   **SSH private key для полдкючения к хосту**
      * ANSIBLE_SSH_PRIVATE_KEY
   
2) Добавьте ваш хост в файл [deploy/inventory.ini](./deploy/ansible/inventory.inil)  также этот хост должен совпадать
с переменной окружения **ANSIBLE_HOST** в [workflows/deploy.yml](.github/workflows/deploy.yml) файле.
4) Сделать push в main

### Если нужно просто запустить сервис, можно использовать docker compose:
```
docker-compose up -d
```

## Пример работы сервиса:
*Запрос на сохранения ссылки c "кличкой" (alias):*
```CURL
curl --request POST \
  --url http://localhost:8082/url/ \
  --header 'Authorization: Basic dnZ2OnBhc3M=' \
  --header 'Content-Type: application/json' \
  --data '{
"url": "https://www.google.com",
"alias": "google"
}'
```
*Ответ от сервиса при успешном сохранении ссылки:*
```
http status OK 200
{
	"status": "OK",
	"alias": "google"
}
```

*Запрос на получение ссылки по alias:*
```CURL
curl --request GET \
  --url http://localhost:8082/google \
  --header 'Authorization: Basic dnZ2OnBhc3M='
```

*Успешный ответ, возвращает найденную ссылку:*
```
http status OK 200

<a href="https://www.google.com">Found</a>.
```

