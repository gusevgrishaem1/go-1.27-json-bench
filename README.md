# Go JSON benchmark lab

Сравнение трёх вариантов JSON на Go 1.27: API v1 со старым движком, API v1 на новом движке и прямой API v2.

| Сервис | API и движок | Порт |
|---|---|---|
| go127-v1-nojsonv2 | encoding/json, старый движок (GOEXPERIMENT=nojsonv2) | 8084 |
| go127-v1 | encoding/json, новый движок с правилами совместимости v1 | 8082 |
| go127-v2 | encoding/json/v2, новый движок с настройками v2 | 8083 |

Старый движок собирается из того же модуля go127-v1. Настройки прямого v2 отличаются от v1, включая порядок ключей map, поэтому сравнение этих вариантов учитывает и правила API.

## Требования

Go 1.27.1, GNU Make, Docker с Docker Compose. Для скрипта бенчмарков в Linux нужен sh, в Windows — PowerShell 7 (pwsh). Команды выполняются из корня репозитория.

## Микробенчмарки

~~~sh
make bench
make bench BENCHTIME=3s COUNT=10
make bench-smoke
~~~

Без Make:

~~~sh
sh ./run-benchmarks.sh 2s 5
~~~

В PowerShell:

~~~powershell
./run-benchmarks.ps1 -Benchtime 2s -Count 5
~~~

По умолчанию выполняются пять повторов с замером 2 секунды на каждый случай. Маленькие объекты содержат 1 элемент, большие — 1000. Encode сериализует Response; decode заполняет новую Request на каждой итерации. Подготовка данных, HTTP и сбор метрик исключены из замера. Результаты включают ns/op, B/op, allocs/op и MB/s.

Скрипты сохраняют результаты и сведения о среде в benchmark-results/. Эта папка и кеш сборки исключены из Git. Сравнивайте результаты на одной машине с одинаковыми параметрами.

## HTTP-сервисы и графики

~~~sh
make up
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
~~~

make up собирает и запускает три сервиса, Prometheus и Grafana. Каждый сервис предоставляет /health, /api/process и /metrics.

- Grafana: http://localhost:3000/d/go-json-v1-v2 — панели длительности encode и decode.
- Prometheus: http://localhost:9090.

Нагрузка k6:

~~~sh
make load PAYLOADS=small VUS=5 DURATION=2m
make load PAYLOADS=large VUS=5 DURATION=2m
make load PAYLOADS=nested ITEMS=500 TEXT_SIZE=256 DURATION=2m
~~~

| Параметр | По умолчанию | Значение |
|---|---|---|
| PAYLOADS | small,medium,large,unicode,nested | Профили через запятую |
| VUS | 5 | Число виртуальных пользователей |
| DURATION | 10m | Длительность нагрузки |
| ITEMS | По профилю | Переопределение числа элементов |
| TEXT_SIZE | 64 | Длина description |
| SLEEP | 0 | Пауза в секундах после обхода сервисов |
| TARGETS | Три сервиса Compose | Адреса HTTP-сервисов через запятую |

Размеры профилей: small — 1 элемент, medium — 100, large — 1000, unicode и nested — 100. k6 передаёт одинаковые данные трём сервисам. Для отдельного сравнения графиков запускайте один профиль за раз: сервисные метрики объединяют все профили.

Графики показывают среднюю длительность JSON-операции на успешный запрос, а не полную HTTP-задержку. Выделенную память оценивайте по B/op и allocs/op микробенчмарков.

Остановка:

~~~sh
make down
~~~

## Проверка Go-модулей

Корневого go.mod нет. В Linux проверяйте каждый вариант отдельно:

~~~sh
(cd go127-v1 && GOEXPERIMENT=jsonv2 go test ./...)
(cd go127-v1 && GOEXPERIMENT=nojsonv2 go test ./...)
(cd go127-v2 && GOEXPERIMENT=jsonv2 go test ./...)
~~~

## Публикация в Git

Папка содержит исходники, конфигурацию запуска и этот README; результаты замеров и материалы статьи не включены.

~~~sh
git init
git add .
git commit -m "Add JSON services and benchmarks"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git
git push -u origin main
~~~

Замените URL адресом своего пустого репозитория.
