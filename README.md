# Go-Expert Tracing

## Como testar a aplicação

Antes de executar o projeto é necessário uma API KEY do https://www.weatherapi.com/ e adicionar na variavel de ambiente `WEATHER_API_KEY` dentro do arquivo `docker-compose.yaml`.

Depois, podemos executar o comando abaixo:

```shell
docker-compose up -d
```

Agora pode fazer requisições para a aplicação:

```shell
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"cep":"01153000"}' \
  http://localhost:8080
```

---
Para acessar o Zipkin e ver o trace use: http://localhost:9411