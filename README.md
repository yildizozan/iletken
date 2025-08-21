# İletken - HTTP Yönlendirici

**İletken**, Go 1.24 ve fasthttp kullanılarak geliştirilmiş, yüksek performanslı HTTP yönlendirici uygulamasıdır.

## Özellikler

- ⚡ **Yüksek Performans**: fasthttp kütüphanesi ile optimize edilmiş
- 📝 **YAML Yapılandırma**: spf13/viper ile esnek yapılandırma
- 🔄 **Esnek Yönlendirme**: 301/302 status kodları ile kalıcı/geçici yönlendirmeler
- 📊 **Structured Logging**: JSON/Text formatında detaylı loglama
- 🛡️ **Graceful Shutdown**: Sinyal yakalama ile güvenli kapatma
- ⚙️ **Kolay Konfigürasyon**: YAML dosyası ile basit kurulum

## Kurulum

```bash
git clone <repository>
cd iletken
go mod download
go build -o iletken
```

## Kullanım

### 1. Yapılandırma

`iletken.yml` dosyasını düzenleyin:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"

redirects:
  - from: "devops.company.com"
    to: "https://hd.company.com/board/13"
  
  - from: "old.example.com"
    to: "https://new.example.com"

logging:
  level: "info"
  format: "json"
```

### 2. Uygulamayı Çalıştırma

```bash
# Varsayılan iletken.yml ile
./iletken

# Farklı config dosyası ile
./iletken -config /path/to/myconfig.yml

# Versiyon bilgisi
./iletken -version
```

### 3. Test

```bash
# HTTP isteği gönder
curl -H "Host: devops.company.com" http://localhost:8080

# Response:
# HTTP/1.1 301 Moved Permanently
# Location: https://hd.company.com/board/13
```

## Yapılandırma Seçenekleri

### Sunucu Ayarları
- `host`: Dinlenecek IP adresi
- `port`: Port numarası
- `read_timeout`: Okuma timeout'u
- `write_timeout`: Yazma timeout'u  
- `idle_timeout`: Boşta kalma timeout'u

### Yönlendirme Kuralları
- `from`: Kaynak host (domain)
- `to`: Hedef URL (tam URL olmalı)

**Not:** Tüm yönlendirmeler 302 (Temporary Redirect) status kodu ile yapılır.

### Loglama
- `level`: Log seviyesi (debug, info, warn, error)
- `format`: Log formatı (json, text)

## Docker ile Çalıştırma

Dockerfile oluşturun:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o iletken

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/iletken .
COPY --from=builder /app/iletken.yml .
EXPOSE 8080
CMD ["./iletken"]
```

```bash
docker build -t iletken .
docker run -p 8080:8080 -v $(pwd)/iletken.yml:/root/iletken.yml iletken
```

## Systemd Servisi

`/etc/systemd/system/iletken.service`:

```ini
[Unit]
Description=İletken HTTP Redirector
After=network.target

[Service]
Type=simple
User=iletken
WorkingDirectory=/opt/iletken
ExecStart=/opt/iletken/iletken -config /opt/iletken/iletken.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable iletken
sudo systemctl start iletken
```

## Performans

fasthttp kullanımı sayesinde yüksek performans:
- Düşük memory allocation
- Hızlı HTTP parsing
- Zero-copy operations
- Connection pooling

## Lisans

MIT License
