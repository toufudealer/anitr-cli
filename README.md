<div align="center">

<h1>anitr-cli</h1>
<h3>Terminalde Türkçe altyazılı anime izleme ve arama aracı 🚀</h3>

<img src="https://raw.githubusercontent.com/xeyossr/anitr-cli/main/assets/anitr-preview.gif" alt="anitr-cli preview" width="300"/>

<p>
  <a href="https://github.com/xeyossr/anitr-cli/releases">
    <img src="https://img.shields.io/github/v/release/xeyossr/anitr-cli?style=for-the-badge&include_prereleases&label=GitHub%20Sürüm">
  </a>
  <a href="https://aur.archlinux.org/packages/anitr-cli">
    <img src="https://img.shields.io/aur/version/anitr-cli?style=for-the-badge&label=AUR">
  </a>
  <a href="https://github.com/mstsecurity/anitr-cli-windows">
    <img src="https://img.shields.io/github/v/release/mstsecurity/anitr-cli-windows?include_prereleases&label=Windows%20Fork&style=for-the-badge">
  </a>
</p>

</div>

---

## 🎬 Özellikler

- **AnimeCix ve OpenAnime Entegrasyonu**: Popüler anime platformlarından hızlı arama ve izleme imkanı.
- **Fansub Seçimi**: OpenAnime üzerinden izlerken favori çeviri grubunuzu seçme özgürlüğü.
- **Çoklu Arayüz Desteği**: Terminal tabanlı TUI ve minimalist grafik arayüz (Rofi UI) seçenekleri.
- **Discord Rich Presence**: İzlediğiniz animeyi Discord profilinizde paylaşarak arkadaşlarınızla etkileşimde kalın.
- **Otomatik Güncelleme Kontrolü**: Uygulama her başlatıldığında yeni sürüm olup olmadığını kontrol eder.

---

## ⚡ Kurulum

### 🐧 Linux

#### Arch tabanlı dağıtımlar (AUR):

```bash
yay -S anitr-cli
```

#### Diğer Linux dağıtımları:

```bash
git clone https://github.com/xeyossr/anitr-cli.git
cd anitr-cli
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
  Depoyu güncelleyip tekrar `make install` komutunu çalıştırın.

---

### 🪟 Windows

Bu proje Linux için geliştirilmiştir. Windows kullanıcıları için [anitr-cli-windows](https://github.com/mstsecurity/anitr-cli-windows) forkunu kullanabilirsiniz.

---

## 🚀 Kullanım

```bash
anitr-cli [bayraklar]
```

**Bayraklar:**

- `--disable-rpc` Discord Rich Presence özelliğini kapatır.
- `--rofi` Rofi arayüzü ile başlatır.
- `--rofi-flags <string>` Rofi için ek parametreler.
- `--version` Sürüm bilgisini gösterir.
- `-h, --help` Yardım menüsü.

---

## 💡 Sorunlar & Katkı

Her türlü hata, öneri veya katkı için [issue](https://github.com/xeyossr/anitr-cli/issues) açabilirsiniz. Katkılarınızı bekliyoruz!

---

## 📜 Lisans

Bu proje [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html) ile lisanslanmıştır. Detaylar için [LICENSE](LICENSE)
