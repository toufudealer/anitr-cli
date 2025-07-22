 # ANITR-CLI

<div align="center">
 
  **Süper Hızlı** bir şekilde anime araması yapabileceğiniz ve istediğiniz animeyi Türkçe altyazılı izleyebileceğiniz terminal aracı 💫

  [![Github_Release](https://img.shields.io/github/v/release/xeyossr/anitr-cli?style=for-the-badge&include_prereleases&label=GitHub%20Release)](https://github.com/xeyossr/anitr-cli/releases) [![AUR](https://img.shields.io/aur/version/anitr-cli?style=for-the-badge)](https://aur.archlinux.org/packages/anitr-cli) [![Windows_Fork](https://img.shields.io/github/v/release/mstsecurity/anitr-cli-windows?include_prereleases&display_name=release&label=Windows%20Fork&style=for-the-badge)](https://github.com/mstsecurity/anitr-cli-windows) 

---

</div>

## 🌟 Özellikler
- **AnimeCix** ve **OpenAnime** desteği: Favori anime sitelerinden animelerinizi izleyin!
- **TUI ve Rofi UI**: Terminal veya minimalist GUI arayüzü ile kullanım.
- **Discord RPC**: İzlediğiniz anime bilgilerini Discord profilinizde gösterin, arkadaşlarınızla paylaşın.

## 💻 Kurulum

### 🐧 Linux Kullanıcıları

Eğer Arch tabanlı bir dağıtım kullanıyorsanız, [AUR](https://aur.archlinux.org/packages/anitr-cli) üzerinden tek bir komut ile indirebilirsiniz:

```bash
yay -S anitr-cli
```

Eğer Arch tabanlı olmayan bir dağıtım kullanıyorsanız, **en son sürümü** indirmek için aşağıdaki komutları kullanabilirsiniz:
```bash
git clone https://github.com/xeyossr/anitr-cli.git
cd anitr-cli
make install
```

> Not: `anitr-cli`'yi kurabilmek için sisteminizde şu paketlerin kurulu olması gerekir:
> 
> **Derleme:** `go`, `git`, `make`   
> **Kullanım (Zorunlu):** `mpv`   
> **İsteğe bağlı: rofi (Rofi arayüzünü kullanmak istiyorsanız)**   
> 
> Bu paketleri aşağıdaki komutlarla sisteminize kurabilirsiniz:
> 
> **Debian / Ubuntu tabanlı sistemler:**
> ```sudo
> sudo apt update
> sudo apt install golang git make mpv rofi
> ```
> 
> **Arch / Manjaro:**
> ```sudo
> sudo pacman -S go git make mpv rofi
> ```
> 
> **Fedora:**
> ```bash
> sudo dnf install golang git make mpv rofi
> ```
> 
> **openSUSE:**
> ```bash
> sudo zypper install go git make mpv rofi
> ```
> Ardından `git clone`, `make install` adımlarını uygulayarak kurulumu tamamlayabilirsiniz.
---

#### Güncelleme

Her çalıştırdığınızda yeni bir güncelleme olup olmadığı denetlenecektir. Eğer güncelleme mevcutsa, şu komutla güncelleyebilirsiniz:

- **AUR** üzerinden kurulum yaptıysanız:
```bash
yay -Sy anitr-cli
```

- **Manuel** kurulum yaptıysanız:
> Eğer manuel kurulum yaptıysanız, güncellemeleri manuel olarak yapmanız gerekmektedir. Alternatif olarak, [anitrcli-update](anitrcli-update) betiğini indirip `/usr/bin` dizinine yerleştirerek otomatik güncelleme yapabilirsiniz.


### 🪟 Windows Kullanıcıları

Bu proje Linux için geliştirilmiştir. **Windows kullanıcıları**, [anitr-cli-windows](https://github.com/mstsecurity/anitr-cli-windows) forkunu kullanabilirler:

> 🔗 [https://github.com/mstsecurity/anitr-cli-windows](https://github.com/mstsecurity/anitr-cli-windows)

## 👾 Kullanım

```bash
💫 Terminalden Türkçe anime izleme aracı

Usage:
  anitr-cli [flags]

Flags:
      --disable-rpc         Discord Rich Presence özelliğini devre dışı bırakır.
  -h, --help                help for anitr-cli
      --rofi                Rofi arayüzü ile başlatır.
      --rofi-flags string   Rofi için flag'ler
      --update              anitr-cli aracını en son sürüme günceller.
      --version             Versiyon bilgisi.
```

## 🚩 Sorunlar ve Katkı

Herhangi bir hata veya geliştirme öneriniz için lütfen bir [**issue**](https://github.com/xeyossr/anitr-cli/issue) açın.

## 📄 Lisans

Bu proje [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html) lisansı ile lisanslanmıştır. Yazılımı bu lisansın koşulları altında kullanmakta, değiştirmekte ve dağıtmakta özgürsünüz. Daha fazla ayrıntı için lütfen [LICENSE](LICENSE) dosyasına bakın.
