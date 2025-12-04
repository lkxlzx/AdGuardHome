&nbsp;
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="doc/adguard_home_darkmode.svg">
    <img alt="AdGuard Home" src="doc/adguard_home_lightmode.svg" width="300px">
  </picture>
</p>
<h3 align="center">您和您设备的隐私保护中心</h3>
<p align="center">
  免费开源、功能强大的全网广告和跟踪器拦截DNS服务器。
</p>
<p align="center">
  <a href="https://adguard.com/">AdGuard.com</a> |
  <a href="https://github.com/AdguardTeam/AdGuardHome/wiki">Wiki</a> |
  <a href="https://reddit.com/r/Adguard">Reddit</a> |
  <a href="https://twitter.com/AdGuard">Twitter</a> |
  <a href="https://t.me/adguard_en">Telegram</a>
  <br/><br/>
  <a href="https://codecov.io/github/AdguardTeam/AdGuardHome?branch=master">
    <img src="https://img.shields.io/codecov/c/github/AdguardTeam/AdGuardHome/master.svg" alt="代码覆盖率"/>
  </a>
  <a href="https://goreportcard.com/report/AdguardTeam/AdGuardHome">
    <img src="https://goreportcard.com/badge/github.com/AdguardTeam/AdGuardHome" alt="Go报告卡"/>
  </a>
  <a href="https://hub.docker.com/r/adguard/adguardhome">
    <img alt="Docker拉取次数" src="https://img.shields.io/docker/pulls/adguard/adguardhome.svg?maxAge=604800"/>
  </a>
  <br/>
  <a href="https://github.com/AdguardTeam/AdGuardHome/releases">
    <img src="https://img.shields.io/github/release/AdguardTeam/AdGuardHome/all.svg" alt="最新版本"/>
  </a>
  <a href="https://snapcraft.io/adguard-home">
    <img alt="adguard-home" src="https://snapcraft.io/adguard-home/badge.svg"/>
  </a>
</p>
<br/>
<p align="center">
  <img src="https://cdn.adtidy.org/public/Adguard/Common/adguard_home.gif" width="800"/>
</p>
<hr/>

AdGuard Home 是一款全网广告和跟踪器拦截软件。设置完成后，它将覆盖您家中的所有设备，无需在客户端安装任何软件。

它作为DNS服务器运行，将跟踪域名重定向到"黑洞"，从而阻止您的设备连接到这些服务器。它基于我们用于公共 [AdGuard DNS] 服务器的软件，两者共享大量代码。

[AdGuard DNS]: https://adguard-dns.io/


- [快速开始](#快速开始)
    - [自动安装（Linux/Unix/MacOS/FreeBSD/OpenBSD）](#自动安装)
    - [其他安装方式](#其他安装方式)
    - [使用指南](#使用指南)
    - [API](#api)
- [与其他解决方案对比](#对比)
    - [与公共AdGuard DNS服务器有何不同？](#对比-adguard-dns)
    - [与Pi-Hole对比](#对比-pi-hole)
    - [与传统广告拦截器对比](#对比-adblock)
    - [已知限制](#对比-限制)
- [从源代码构建](#从源代码构建)
    - [前置要求](#前置要求)
    - [构建](#构建)
- [贡献](#贡献)
    - [测试不稳定版本](#测试不稳定版本)
    - [报告问题](#报告问题)
    - [帮助翻译](#帮助翻译)
    - [其他](#其他贡献)
- [使用AdGuard Home的项目](#使用项目)
- [致谢](#致谢)
- [隐私](#隐私)

## <a href="#快速开始" id="快速开始" name="快速开始">快速开始</a>

### <a href="#自动安装" id="自动安装" name="自动安装">自动安装（Linux/Unix/MacOS/FreeBSD/OpenBSD）</a>

使用 `curl` 安装，运行以下命令：

```sh
curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

使用 `wget` 安装，运行以下命令：

```sh
wget --no-verbose -O - https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

使用 `fetch` 安装，运行以下命令：

```sh
fetch -o - https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

脚本还接受一些选项：

- `-c <channel>` 使用指定的更新频道；
- `-r` 重新安装AdGuard Home；
- `-u` 卸载AdGuard Home；
- `-v` 详细输出。

注意：选项 `-r` 和 `-u` 互斥。

### <a href="#其他安装方式" id="其他安装方式" name="其他安装方式">其他安装方式</a>

#### <a href="#手动安装" id="手动安装" name="手动安装">手动安装</a>

请阅读我们Wiki上的 **[快速开始][wiki-start]** 文章，了解如何手动安装AdGuard Home，以及如何配置您的设备使用它。

#### <a href="#docker" id="docker" name="docker">Docker</a>

您可以使用我们在 [Docker Hub] 上的官方Docker镜像。

#### <a href="#snap-store" id="snap-store" name="snap-store">Snap Store</a>

如果您运行的是 **Linux**，有一种安全简便的方式安装AdGuard Home：从 [Snap Store] 获取。

[Docker Hub]: https://hub.docker.com/r/adguard/adguardhome
[Snap Store]: https://snapcraft.io/adguard-home
[wiki-start]: https://adguard-dns.io/kb/adguard-home/getting-started/

### <a href="#使用指南" id="使用指南" name="使用指南">使用指南</a>

查看我们的 [Wiki][wiki]。

[wiki]: https://github.com/AdguardTeam/AdGuardHome/wiki

### <a href="#api" id="api" name="api">API</a>

如果您想与AdGuard Home集成，可以使用我们的 [REST API][openapi]。或者，您可以使用这个 [Python客户端][pyclient]，它用于构建 [AdGuard Home Hass.io插件][hassio]。

[hassio]:   https://www.home-assistant.io/integrations/adguard/
[openapi]:  https://github.com/AdguardTeam/AdGuardHome/tree/master/openapi
[pyclient]: https://pypi.org/project/adguardhome/


## <a href="#对比" id="对比" name="对比">与其他解决方案对比</a>

### <a href="#对比-adguard-dns" id="对比-adguard-dns" name="对比-adguard-dns">与公共AdGuard DNS服务器有何不同？</a>

运行您自己的AdGuard Home服务器可以做的事情远比使用公共DNS服务器多得多。这是完全不同的级别。亲自看看：

- 选择服务器具体拦截和允许什么。

- 监控您的网络活动。

- 添加您自己的自定义过滤规则。

- **最重要的是，这是您自己的服务器，您是唯一的控制者。**

### <a href="#对比-pi-hole" id="对比-pi-hole" name="对比-pi-hole">与Pi-Hole对比</a>

目前，AdGuard Home与Pi-Hole有很多共同点。两者都使用所谓的"DNS沉洞"方法拦截广告和跟踪器，并且都允许自定义拦截内容。

> [!NOTE]
> 我们不会止步于此。DNS沉洞不是一个坏的起点，但这只是开始。

AdGuard Home开箱即用提供了许多功能，无需安装和配置额外的软件。我们希望它足够简单，即使是普通用户也能以最小的努力完成设置。

> [!NOTE]
> 列出的一些功能可以通过安装额外软件或手动使用SSH终端重新配置Pi-Hole的某个组件来添加到Pi-Hole。但是，在我们看来，这不能合理地算作Pi-Hole的功能。

| 功能                                                                 | AdGuard&nbsp;Home | Pi-Hole                                                   |
|----------------------------------------------------------------------|-------------------|-----------------------------------------------------------|
| 拦截广告和跟踪器                                                     | ✅                | ✅                                                        |
| 自定义拦截列表                                                       | ✅                | ✅                                                        |
| 内置DHCP服务器                                                       | ✅                | ✅                                                        |
| 管理界面HTTPS                                                        | ✅                | 有，但需要手动配置lighttpd                                |
| 加密DNS上游服务器（DNS-over-HTTPS、DNS-over-TLS、DNSCrypt）         | ✅                | ❌（需要额外软件）                                        |
| 跨平台                                                               | ✅                | ❌（非原生，仅通过Docker）                                |
| 作为DNS-over-HTTPS或DNS-over-TLS服务器运行                          | ✅                | ❌（需要额外软件）                                        |
| 拦截钓鱼和恶意软件域名                                               | ✅                | ❌（需要非默认拦截列表）                                  |
| 家长控制（拦截成人域名）                                             | ✅                | ❌（需要非默认拦截列表）                                  |
| 在搜索引擎上强制安全搜索                                             | ✅                | ❌                                                        |
| 按客户端（设备）配置                                                 | ✅                | ✅                                                        |
| 访问设置（选择谁可以使用AGH DNS）                                    | ✅                | ❌                                                        |
| [无需root权限运行][wiki-noroot]                                      | ✅                | ❌                                                        |

[wiki-noroot]: https://adguard-dns.io/kb/adguard-home/getting-started/#running-without-superuser

### <a href="#对比-adblock" id="对比-adblock" name="对比-adblock">与传统广告拦截器对比</a>

这取决于情况。

DNS沉洞能够拦截很大比例的广告，但它缺乏传统广告拦截器的灵活性和强大功能。您可以通过阅读 [这篇文章][blog-adaway] 来了解这些方法之间的区别，该文章比较了AdGuard for Android（传统广告拦截器）与hosts级别的广告拦截器（其功能几乎与基于DNS的拦截器相同）。这种保护级别对某些用户来说已经足够。

此外，使用基于DNS的拦截器可以帮助拦截其他类型设备上的广告、跟踪和分析请求，例如智能电视、智能音箱或其他类型的物联网设备（您无法在这些设备上安装传统广告拦截器）。

### <a href="#对比-限制" id="对比-限制" name="对比-限制">已知限制</a>

以下是一些DNS级别拦截器无法拦截的示例：

- YouTube、Twitch广告；

- Facebook、Twitter、Instagram赞助帖子。

本质上，任何与内容共享域名的广告都无法被DNS级别拦截器拦截。

未来有机会处理这个问题吗？DNS永远不足以做到这一点。我们唯一的选择是使用内容拦截代理，就像我们在独立的AdGuard应用程序中所做的那样。我们 [计划][issue-1228] 在未来为AdGuard Home带来此功能支持。不幸的是，即使在这种情况下，仍然会有一些情况这还不够，或者需要相当复杂的配置。

[blog-adaway]: https://adguard.com/blog/adguard-vs-adaway-dns66.html
[issue-1228]:  https://github.com/AdguardTeam/AdGuardHome/issues/1228


## <a href="#从源代码构建" id="从源代码构建" name="从源代码构建">从源代码构建</a>

### <a href="#前置要求" id="前置要求" name="前置要求">前置要求</a>

运行 `make init` 来准备开发环境。

构建AdGuard Home需要：

- [Go](https://golang.org/dl/) v1.25或更高版本；
- [Node.js](https://nodejs.org/en/download/) v24.10.0或更高版本；
- [npm](https://www.npmjs.com/) v10.8或更高版本；

### <a href="#构建" id="构建" name="构建">构建</a>

打开终端并执行以下命令：

```sh
git clone https://github.com/AdguardTeam/AdGuardHome
cd AdGuardHome
make
```

> [!WARNING]
> 目前不支持非标准的 `-j` 标志，因此使用 `make -j 4` 构建或将 `MAKEFLAGS` 设置为包含例如 `-j 4` 可能会破坏构建。如果您确实设置了 `MAKEFLAGS`，并且不想更改它，可以通过运行 `make -j 1` 来覆盖它。

查看 [`Makefile`][src-makefile] 了解其他命令。

#### <a href="#跨平台构建" id="跨平台构建" name="跨平台构建">为不同平台构建</a>

您可以为Go支持的任何OS/ARCH构建AdGuard Home。为此，在运行 `make` 时将 `GOOS` 和 `GOARCH` 环境变量指定为宏。

例如：

```sh
env GOOS='linux' GOARCH='arm64' make
```

或：

```sh
make GOOS='linux' GOARCH='arm64'
```

#### <a href="#准备发布版本" id="准备发布版本" name="准备发布版本">准备发布版本</a>

您需要 [`snapcraft`] 来准备发布版本。安装后，运行以下命令：

```sh
make build-release CHANNEL='...' VERSION='...'
```

查看 [`build-release` 目标文档][targ-release]。

#### <a href="#docker镜像" id="docker镜像" name="docker镜像">Docker镜像</a>

运行 `make build-docker` 在本地构建Docker镜像（我们发布到DockerHub的镜像）。请注意，我们使用 [Docker Buildx][buildx] 来构建官方镜像。

使用这些构建之前，您可能需要准备：

- （仅限Linux）安装Qemu：

  ```sh
  docker run --rm --privileged multiarch/qemu-user-static --reset -p yes --credential yes
  ```

- 准备构建器：

  ```sh
  docker buildx create --name buildx-builder --driver docker-container --use
  ```

查看 [`build-docker` 目标文档][targ-docker]。

#### <a href="#调试前端" id="调试前端" name="调试前端">调试前端</a>

当您需要调试前端而不是每次都重新编译生产版本时，例如检查标签在表单上的外观，您可以在开发环境中运行前端构建。

1. 在单独的终端中运行：

   ```sh
   ( cd ./client/ && env NODE_ENV='development' npm run watch )
   ```

2. 使用 `--local-frontend` 标志运行 `AdGuardHome` 二进制文件，这会指示AdGuard Home忽略内置的前端文件，使用 `./build/` 目录中的文件。

3. 现在您在 `./client/` 目录中所做的任何更改都应该被重新编译并在Web UI上可用。确保禁用浏览器缓存，以确保您实际获得重新编译的版本。

[`snapcraft`]:  https://snapcraft.io/
[buildx]:       https://docs.docker.com/buildx/working-with-buildx/
[src-makefile]: https://github.com/AdguardTeam/AdGuardHome/blob/master/Makefile
[targ-docker]:  https://github.com/AdguardTeam/AdGuardHome/tree/master/scripts#build-dockersh-build-a-multi-architecture-docker-image
[targ-release]: https://github.com/AdguardTeam/AdGuardHome/tree/master/scripts#build-releasesh-build-a-release-for-all-platforms

#### <a href="#e2e前端测试" id="e2e前端测试" name="e2e前端测试">端到端（E2E）前端测试</a>

AdGuard Home使用 [Playwright](https://playwright.dev) 进行E2E测试。测试位于 `tests/e2e`。

**运行测试：**
- `npm run test:e2e` – 运行所有测试（无头模式）。
- `npm run test:e2e:interactive` – 交互式运行测试。
- `npm run test:e2e:debug` – 调试模式运行测试。
- `npm run test:e2e:codegen` – 生成新的测试代码。

**设置：**
1. 运行 `npm install` 安装依赖。
2. 运行 `npx playwright install` 设置所需的浏览器。

> **警告：** Playwright将下载并安装自己的浏览器二进制文件用于测试，这可能与系统上安装的浏览器不同。


## <a href="#贡献" id="贡献" name="贡献">贡献</a>

欢迎您fork此仓库，进行更改并 [提交pull request][pr]。但请确保遵循我们的 [代码指南][guide]。

请注意，我们不期望人们同时为程序的UI和后端部分做出贡献。理想情况下，后端部分首先实现，即配置、API和功能本身。UI部分可以稍后由不同的人在不同的pull request中实现。

[guide]: https://github.com/AdguardTeam/CodeGuidelines/
[pr]:    https://github.com/AdguardTeam/AdGuardHome/pulls

### <a href="#测试不稳定版本" id="测试不稳定版本" name="测试不稳定版本">测试不稳定版本</a>

您可以使用两个更新频道：

- `beta`：AdGuard Home的beta版本。相对稳定的版本，通常每两周或更频繁发布。

- `edge`：来自开发分支的最新版本AdGuard Home。每天都会向此频道推送新更新。

有三种方式可以安装不稳定版本：

1. [Snap Store]：查找 `beta` 和 `edge` 频道。

2. [Docker Hub]：查找 `beta` 和 `edge` 标签。

3. 独立构建。使用自动安装脚本或在 [Wiki][wiki-platf] 上查找可用的构建。

   安装beta版本的脚本：

   ```sh
   curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -c beta
   ```

   安装edge版本的脚本：

   ```sh
   curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -c edge
   ```

[wiki-platf]: https://github.com/AdguardTeam/AdGuardHome/wiki/Platforms

### <a href="#报告问题" id="报告问题" name="报告问题">报告问题</a>

如果您遇到任何问题或有建议，请前往 [此页面][iss] 并点击"New issue"按钮。请仔细遵循问题表单中的说明，不要忘记首先搜索重复项。

[iss]: https://github.com/AdguardTeam/AdGuardHome/issues

### <a href="#帮助翻译" id="帮助翻译" name="帮助翻译">帮助翻译</a>

如果您想帮助翻译AdGuard Home，请在我们的 [知识库][kb-trans] 中了解更多关于翻译AdGuard产品的信息。您可以在 [CrowdIn上的AdGuardHome项目][crowdin] 中贡献。

[crowdin]:  https://crowdin.com/project/adguard-applications/en#/adguard-home
[kb-trans]: https://kb.adguard.com/en/general/adguard-translations

### <a href="#其他贡献" id="其他贡献" name="其他贡献">其他贡献</a>

您可以通过 [查找标记为][iss-help] `help wanted` 的问题，询问该问题是否可以认领，并发送修复bug或实现功能的PR来做出贡献。

[iss-help]: https://github.com/AdguardTeam/AdGuardHome/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22


## <a href="#使用项目" id="使用项目" name="使用项目">使用AdGuard Home的项目</a>

请注意，这些项目不隶属于AdGuard，而是由第三方开发者和粉丝制作的。

- [AdGuard Home Remote](https://apps.apple.com/app/apple-store/id1543143740)：[Joost](https://rocketscience-it.nl/) 开发的iOS应用。

- [Python库](https://github.com/frenck/python-adguardhome)：[@frenck](https://github.com/frenck) 开发。

- [Home Assistant插件](https://github.com/hassio-addons/addon-adguard-home)：[@frenck](https://github.com/frenck) 开发。

- [OpenWrt LUCI应用](https://github.com/kongfl888/luci-app-adguardhome)：[@kongfl888](https://github.com/kongfl888) 开发（最初由 [@rufengsuixing](https://github.com/rufengsuixing) 开发）。

- [AdGuardHome同步](https://github.com/bakito/adguardhome-sync)：[@bakito](https://github.com/bakito) 开发。

- [基于终端的实时流量监控和统计](https://github.com/Lissy93/AdGuardian-Term)：[@Lissy93](https://github.com/Lissy93) 开发。

- [GLInet路由器上的AdGuard Home](https://forum.gl-inet.com/t/adguardhome-on-gl-routers/10664)：[Gl-Inet](https://gl-inet.com/) 开发。

- [Cloudron应用](https://git.cloudron.io/cloudron/adguard-home-app)：[@gramakri](https://github.com/gramakri) 开发。

- [Asuswrt-Merlin-AdGuardHome-Installer](https://github.com/jumpsmm7/Asuswrt-Merlin-AdGuardHome-Installer)：[@jumpsmm7](https://github.com/jumpsmm7) 又名 [@SomeWhereOverTheRainBow](https://www.snbforums.com/members/somewhereovertherainbow.64179/) 开发。

- [Node.js库](https://github.com/Andrea055/AdguardHomeAPI)：[@Andrea055](https://github.com/Andrea055/) 开发。

- [浏览器扩展](https://github.com/satheshshiva/Adguard-Home-Browser-Ext)：[@satheshshiva](https://github.com/satheshshiva/) 开发。

- [Zabbix模板](https://github.com/diasdmhub/AdGuard_Home_Zabbix_Template)：[@diasdmhub](https://github.com/diasdmhub) 开发。

- [Chocolatey包](https://community.chocolatey.org/packages/adguardhome/)：[niks255](https://community.chocolatey.org/profiles/niks255) 开发。

## <a href="#致谢" id="致谢" name="致谢">致谢</a>

如果没有以下软件，这个软件是不可能实现的：

- [Go](https://golang.org/dl/) 及其库：
    - [gcache](https://github.com/bluele/gcache)
    - [miekg's dns](https://github.com/miekg/dns)
    - [go-yaml](https://github.com/go-yaml/yaml)
    - [service](https://godoc.org/github.com/kardianos/service)
    - [dnsproxy](https://github.com/AdguardTeam/dnsproxy)
    - [urlfilter](https://github.com/AdguardTeam/urlfilter)
- [Node.js](https://nodejs.org/) 及其库：
    - [React.js](https://reactjs.org)
    - [Tabler](https://github.com/tabler/tabler)
    - 以及更多Node.js包。
- [whotracks.me数据](https://github.com/cliqz-oss/whotracks.me)

您可能已经看到之前这里提到了 [CoreDNS]，但我们已经停止在AdGuard Home中使用它。

有关使用的所有Node.js包的完整列表，请查看 [`client/package.json`][src-packagejson] 文件。

[CoreDNS]:         https://coredns.io
[src-packagejson]: https://github.com/AdguardTeam/AdGuardHome/blob/master/client/package.json

## <a href="#隐私" id="隐私" name="隐私">隐私</a>

我们的主要理念是您应该掌控自己的数据。因此，AdGuard Home不收集任何使用统计信息，除非您配置它这样做，否则不使用任何Web服务。另请参阅 [完整隐私政策][privacy]，其中包含AdGuard Home *理论上可能发送* 的每一项内容。

[privacy]: https://adguard.com/en/privacy/home.html

