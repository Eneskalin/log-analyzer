# Log Analyzer

Go dilinde geliştirilmiş, Bubble Tea (TUI) çerçevesini kullanan yüksek performanslı, gerçek zamanlı günlük analizi ve izleme aracı. Bu uygulama, karmaşık günlük dosyalarını ayrıştırmak, güvenlik tehditlerini tanımlamak ve canlı sistem izleme sağlamak için tasarlanmıştır.## Technical Stack

- **Programlama Dili:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
## Özellikler

- **Eş Zamanlı Takip (Tailing):** Yeni girişler için anlık uyarı bildirimleri ile birden fazla günlük dosyasını aynı anda canlı olarak takip edin.
- **Otonom Log Analizi** Toplam satır sayısı, olay eşleştirme ve önem derecesi istatistikleri (Kritik, Hata, Bilgi) dahil olmak üzere kapsamlı özetler oluşturur..
- **Güvemlik :**  Özelleştirilebilir kurallara dayalı olarak SQL Enjeksiyonu, XSS ve yetkisiz erişim girişimleri gibi kalıpları tespit etmek için tasarlanmıştır.
- **Dışa Aktarma** Analiz sonuçlarını daha ayrıntılı inceleme veya raporlama için CSV formatına aktarma özelliği.





## Proje Mimarisi

- `/ui`: TUI mantığını, menü yönetimini ve görünüm oluşturmayı içerir.
- `/helpers`: Günlük okuma, desen eşleştirme ve özet oluşturma mantığı.
- `/models`: Günlükler, kurallar ve uygulama durumu için dahili veri yapıları.
- `/config`: Günlük yolları ve algılama kuralları için JSON yapılandırma dosyaları.


## Özelleştirme

 `/config` klasörü üzerinden izlenecek logların ve pathlar ve kurallar ile  özelleşterilebilmektedir.


**Örnek:**
```json
{
  "logs": {
    "Web-Server": "logs/nginx/access.log",
    "System-Auth": "logs/auth.log",
    "Firewall": "logs/ufw.log"
  }
}
```




### Kurulum

 Repoyu klonla:
   ```bash
   git clone https://github.com/eneskalin/log-analyzer.git
   ```
   Projeyi ayağa kaldır
   ```bash
   cd log-analyzer
   docker-compose up --build
   
  ```
  Yeni bir termiinalde çalıştır
  ```bash
  docker-compose run log-analyzer
```


## Ekran Görüntüleri


![Uygulama Ekran Görüntüsü](https://i.imgur.com/wRZP9lX.gif)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/8gEN0jJ.gif)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/7uDua0m.png)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/PdnxTqv.gif)


