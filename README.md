<div align="center">

<h1>anitr-cli</h1>
<h3>Terminalde Türkçe altyazılı anime izleme ve arama aracı 🚀</h3>

<img src="https://raw.githubusercontent.com/xeyossr/anitr-cli/main/assets/anitr-preview.gif" alt="anitr-cli preview" width="600"/>

<p>
  
[![Lisans: GPL3](https://img.shields.io/github/license/xeyossr/anitr-cli?style=for-the-badge&logo=opensourceinitiative&logoColor=white&label=Lisans)](https://github.com/xeyossr/anitr-cli/blob/main/LICENSE)
[![Go Versiyon](https://img.shields.io/badge/Go-1.24+-blue?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/xeyossr/anitr-cli?style=for-the-badge&logo=github&logoColor=white&label=Son%20Sürüm)](https://github.com/xeyossr/anitr-cli/releases/latest)
[![AUR](https://img.shields.io/aur/version/anitr-cli?style=for-the-badge&logo=archlinux&logoColor=white&label=AUR)](https://aur.archlinux.org/packages/anitr-cli)
    
</p>

</div>

---

## 🎬 Özellikler

- **AnimeCix ve OpenAnime Entegrasyonu**: Popüler anime platformlarından hızlı arama ve izleme imkanı.
- **Fansub Seçimi**: OpenAnime üzerinden izlerken favori çeviri grubunuzu seçme özgürlüğü.
- **Çoklu Arayüz Desteği**: Terminal tabanlı TUI ve minimalist grafik arayüz (Rofi UI) seçenekleri.
- **Discord Rich Presence**: İzlediğiniz animeyi Discord profilinizde paylaşın.
- **Otomatik Güncelleme Kontrolü**: Uygulama her başlatıldığında yeni sürüm olup olmadığını kontrol eder.

---

## ⚡ Kurulum

### 🐧 Arch tabanlı dağıtımlar (AUR):

```bash
yay -S anitr-cli
```
ya da
```bash
paru -S anitr-cli
```

### 🐧 Diğer Linux dağıtımları:

```bash
curl -sS https://raw.githubusercontent.com/xeyossr/anitr-cli/main/install.sh | bash
```
ya da
```bash
git clone https://github.com/xeyossr/anitr-cli.git
cd anitr-cli  
git fetch --tags
make install
```

> **Gereksinimler:**  
> Derleme: `go`, `git`, `make`  
> Kullanım: `mpv`  
> İsteğe bağlı: `rofi` (Rofi arayüzü için)

**Paketleri yüklemek için:**

- **Debian/Ubuntu:**
  ```bash
  sudo apt update
  sudo apt install golang git make mpv rofi
  ```
- **Arch/Manjaro:**
  ```bash
  sudo pacman -S go git make mpv rofi
  ```
- **Fedora:**
  ```bash
  sudo dnf install golang git make mpv rofi
  ```
- **openSUSE:**
  ```bash
  sudo zypper install go git make mpv rofi
  ```

---

### 🔄 Güncelleme

- **AUR ile kurduysanız:**
  ```bash
  yay -Sy anitr-cli
  ```
- **Manuel kurulum yaptıysanız:**  
  ```bash
  curl -sS https://raw.githubusercontent.com/xeyossr/anitr-cli/main/install.sh | bash
  ```

---

## 🚀 Kullanım

```bash
anitr-cli [alt komut] [bayraklar]
```

### Bayraklar
- `--disable-rpc`  
  Discord Rich Presence özelliğini kapatır.
- `--version`, `-v`  
  Sürüm bilgisini gösterir.
- `--help`, `-h`  
  Yardım menüsünü gösterir.
- `--rofi`  
  **[Kullanımdan kaldırıldı]** Yerine `rofi` alt komutunu kullanın.

### Alt Komutlar
- `rofi`  
  Rofi arayüzü ile başlatır.  
  - `-f`, `--rofi-flags`  
    Rofi’ye özel parametreler (örn: `--rofi-flags="-theme mytheme"`).

- `tui`  
  Terminal arayüzü ile başlatır.

## 💡 Sorunlar & Katkı

Her türlü hata, öneri veya katkı için [issue](https://github.com/xeyossr/anitr-cli/issues) açabilirsiniz. Katkılarınızı bekliyoruz!

---

## 📜 Lisans

Bu proje [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html) ile lisanslanmıştır. Detaylar için [LICENSE](LICENSE)
