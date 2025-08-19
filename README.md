<div align="center">

<h1>anitr-cli</h1>
<h3>Terminalde Türkçe altyazılı anime arama ve izleme aracı 🚀</h3>

<img src="https://raw.githubusercontent.com/xeyossr/anitr-cli/main/assets/anitr-preview.gif" alt="anitr-cli preview" width="600"/>

<p>
  
[![Lisans: GPL3](https://img.shields.io/github/license/xeyossr/anitr-cli?style=for-the-badge&logo=opensourceinitiative&logoColor=white&label=Lisans)](https://github.com/xeyossr/anitr-cli/blob/main/LICENSE)
[![Go Versiyon](https://img.shields.io/badge/Go-1.24+-blue?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/xeyossr/anitr-cli?style=for-the-badge&logo=github&logoColor=white&label=Son%20Sürüm)](https://github.com/xeyossr/anitr-cli/releases/latest)
[![AUR](https://img.shields.io/aur/version/anitr-cli?style=for-the-badge&logo=archlinux&logoColor=white&label=AUR)](https://aur.archlinux.org/packages/anitr-cli)
    
</p>

</div>

> [!NOTE]
> Bu proje, [xeyossr/anitr-cli](https://github.com/xeyossr/anitr-cli) projesinin bir fork'udur.

--- 

## 🎬 Özellikler

- **Cross-Platform**: Bu fork özellikle Windows için geliştirilmiştir.
- **AnimeCix ve OpenAnime Entegrasyonu**: Popüler anime platformlarından hızlı arama ve izleme imkanı.
- **Fansub Seçimi**: OpenAnime üzerinden izlerken favori çeviri grubunuzu seçme özgürlüğü.
- **Çoklu Arayüz Desteği**: Terminal tabanlı TUI ve minimalist grafik arayüz (Rofi UI) seçenekleri.
- **Discord Rich Presence**: İzlediğiniz animeyi Discord profilinizde paylaşın.
- **Otomatik Güncelleme Kontrolü**: Uygulama her başlatıldığında yeni sürüm olup olmadığını kontrol eder.

--- 

## ⚡ Kurulum

## 🐧 Linux

### Arch tabanlı dağıtımlar (AUR):

```bash
yay -S anitr-cli
```
ya da
```bash
paru -S anitr-cli
```

### Diğer Linux dağıtımları:

```bash
curl -sS https://raw.githubusercontent.com/xeyossr/anitr-cli/main/install.sh | bash
```
ya da
```bash
git clone https://github.com/xeyossr/anitr-cli.git
cd anitr-cli  
git fetch --tags
make install-linux
```

> **Gereksinimler:**  
> Derleme: `go`, `git`, `make`  
> Kullanım: `vlc`  
> İsteğe bağlı: `rofi` (Rofi arayüzü için)

**Paketleri yüklemek için:**

- **Debian/Ubuntu:**
  ```bash
  sudo apt update
  sudo apt install golang git make vlc rofi
  ```
- **Arch/Manjaro:**
  ```bash
  sudo pacman -S go git make vlc rofi
  ```
- **Fedora:**
  ```bash
  sudo dnf install golang git make vlc rofi
  ```
- **openSUSE:**
  ```bash
  sudo zypper install go git make vlc rofi
  ```

## 🪟 Windows

> [!IMPORTANT]
> Bu fork, özellikle Windows işletim sistemi için geliştirilmiştir. Linux veya macOS desteği sağlanmamaktadır.

> [!NOTE]
> Windows sürümünde GUI bulunmaz, yalnızca TUI ile çalışır.

1.  **VLC Media Player Kurulumu:**
    *   `anitr-cli` animeleri oynatmak için [**VLC Media Player**](https://www.videolan.org/) kullanır. Lütfen sisteminizde VLC'nin kurulu olduğundan emin olun.
    *   **VLC'nin PATH Ortam Değişkeninde Olması:** `anitr-cli` varsayılan olarak VLC'yi sisteminizin `PATH` ortam değişkeninde arar. `PATH`, işletim sisteminizin çalıştırılabilir programları aradığı dizinlerin bir listesidir.
        *   **VLC'nin PATH'te olup olmadığını kontrol etmek için:** Komut İstemi'ni (CMD) veya PowerShell'i açın ve `vlc --version` yazın. Eğer VLC sürüm bilgisi görünüyorsa, VLC PATH'inizdedir. Aksi takdirde "komut bulunamadı" gibi bir hata alırsınız.
        *   **VLC'yi PATH'e eklemek için (genel rehber):** 
            *   VLC'nin kurulu olduğu dizini (örneğin, `C:\Program Files\VideoLAN\VLC`) kopyalayın.
            *   Windows Arama çubuğuna "ortam değişkenleri" yazın ve "Sistem ortam değişkenlerini düzenleyin" seçeneğini açın.
            *   "Ortam Değişkenleri..." düğmesine tıklayın.
            *   "Sistem değişkenleri" altında `Path` değişkenini bulun, seçin ve "Düzenle..." düğmesine tıklayın.
            *   "Yeni"ye tıklayın ve VLC'nin kurulu olduğu dizini yapıştırın. Tamam'a tıklayarak tüm pencereleri kapatın.
            *   **Önemli:** Değişikliklerin etkili olması için yeni bir Komut İstemi veya PowerShell penceresi açmanız gerekebilir.
    *   **Alternatif (VLC PATH'te değilse):** Eğer VLC'yi PATH'e eklemek istemiyorsanız veya birden fazla VLC sürümünüz varsa, `anitr-cli` uygulamasının kaynak kodunu düzenleyerek VLC'nin tam yolunu belirtebilirsiniz.
        *   `anitr-cli` projesini indirin ve bir metin düzenleyici ile `main.go` dosyasını açın.
        *   `playAnimeLoop` fonksiyonu içinde şu satırı bulun:
            ```go
            VLCPath: "", // Assuming VLC is in PATH, otherwise specify
            ```
        *   Bu satırı VLC yürütülebilir dosyanızın tam yolu ile değiştirin. Örneğin:
            ```go
            VLCPath: "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe", // VLC'nin tam yolu
            ```
        *   Değişikliği kaydettikten sonra, `anitr-cli` projesinin ana dizininde Komut İstemi'ni açın ve uygulamayı yeniden derlemek için `go build` komutunu çalıştırın.

2.  [Releases](https://github.com/xeyossr/anitr-cli/releases) sayfasından `anitr-cli.exe` indirin.
3.  `C:\Program Files\anitr-cli` klasörünü oluşturun.
4.  `anitr-cli.exe` dosyasını bu klasöre taşıyın.
5.  PATH’e `C:\Program Files\anitr-cli` ekleyin.

Artık **cmd** veya **PowerShell** içinde anitr-cli çalıştırabilirsiniz.

## 💻 MacOS

> [!WARNING]
> Mac cihazım olmadığından dolayı **anitr-cli** MacOS üzerinde test edilmedi. 
> Ancak, Linux'ta kullanılan yöntemlerle kurulup çalışması oldukça muhtemeldir. Herhangi bir sorunla karşılaşırsanız lütfen [**issue**](https://github.com/xeyossr/anitr-cli/issues) açınız.

**Kurulum (Manuel)**:
```bash
git clone https://github.com/xeyossr/anitr-cli.git
cd anitr-cli
git fetch --tags
make install-macos
```

--- 

## 🚀 Kullanım

```bash
anitr-cli [alt komut] [bayraklar]
```

Bayraklar:   
  `--disable-rpc`         Discord Rich Presence özelliğini kapatır   
  `--version`, `-v`       Sürüm bilgisini gösterir   
  `--help`, `-h`          Yardım menüsünü gösterir   
  `--rofi`                **[Kullanımdan kaldırıldı]** Yerine 'rofi' alt komutunu kullanın (Sadece Linux)  

Alt komutlar: (Sadece Linux)
  `rofi`                  Rofi arayüzü ile başlatır
    `-f`, `--rofi-flags`  Rofi’ye özel parametreler (örn: `--rofi-flags="-theme mytheme"`)   
  `tui`                   Terminal arayüzü ile başlatır   

--- 

## 💡 Sorunlar & Katkı

Her türlü hata, öneri veya katkı için [issue](https://github.com/xeyossr/anitr-cli/issues) açabilirsiniz. Katkılarınızı bekliyoruz!

--- 

## 📜 Lisans

Bu proje [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html) ile lisanslanmıştır. Detaylar için [LICENSE](LICENSE)