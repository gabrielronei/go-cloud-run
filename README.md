# Go Cloud Run - CEP Weather Service

Serviço em Go que recebe um CEP, identifica a cidade via ViaCEP e retorna as temperaturas em Celsius, Fahrenheit e Kelvin via WeatherAPI. Hospedado no Google Cloud Run.

## URL do Cloud Run

https://go-cloud-run-242000668090.europe-west1.run.app

## Endpoint

```
GET /{cep}
```

### Exemplo

```bash
curl https://go-cloud-run-242000668090.europe-west1.run.app/01310100
```

### Respostas

**200 OK**
```json
{ "temp_C": 25.0, "temp_F": 77.0, "temp_K": 298.0 }
```

**422 Unprocessable Entity** — CEP inválido
```
invalid zipcode
```

**404 Not Found** — CEP não encontrado
```
can not find zipcode
```

## Rodando localmente com Docker

```bash
docker build -t cep-weather .
docker run -p 8080:8080 -e WEATHER_API_KEY=sua_key cep-weather
```

Acesse: `http://localhost:8080/01310100`

## Rodando os testes

```bash
go test ./... -v
```

Os testes cobrem:
- Conversões de temperatura (Celsius → Fahrenheit e Kelvin)
- Handler HTTP com mock das APIs externas (sucesso, CEP inválido, CEP não encontrado)

## Variáveis de Ambiente

| Variável | Descrição |
|---|---|
| `WEATHER_API_KEY` | Chave da [WeatherAPI](https://www.weatherapi.com/) |
| `PORT` | Porta do servidor (padrão: `8080`) |

## Deploy no Cloud Run

```bash
gcloud run deploy cep-weather \
  --source . \
  --region europe-west1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_key
```
## Resultado
- Imagem do projeto rodando certinho
<img width="1443" height="138" alt="image" src="https://github.com/user-attachments/assets/6c394735-10d2-4b7d-ba89-5152d668748b" />

