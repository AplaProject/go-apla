<p align="center">
  <img src="https://i.imgur.com/4lMw23m.png" width="500">
</p>
<br>

[![Throughput Graph](https://graphs.waffle.io/GenesisKernel/go-genesis/throughput.svg)](https://waffle.io/GenesisKernel/go-genesis/metrics/throughput)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/GenesisKernel/go-genesis)](https://goreportcard.com/report/github.com/GenesisKernel/go-genesis)
[![Build Status](https://travis-ci.org/GenesisKernel/go-genesis.svg?branch=master)](https://travis-ci.org/GenesisKernel/go-genesis)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg?style=flat)](http://genesiskernel.readthedocs.io/en/latest/)
[![](https://tokei.rs/b1/github/GenesisKernel/go-genesis)](https://github.com/GenesisKernel/go-genesis)
![](https://reposs.herokuapp.com/?path=GenesisKernel/go-genesis&style=flat)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/GenesisKernel/go-genesis)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/GenesisKernel?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)


<p align="center">
  <a href="README.md">EN</a> | CN | <a href="README-ES.md">ES</a> | <a href="README-RU.md">RU</a>
</p>


## 目录

- [引言](#%E5%BC%95%E8%A8%80)
- [Genesis的特点？](#%E7%9A%84%E7%89%B9%E7%82%B9)
- [获得自己的githubUser代币！](#%E8%8E%B7%E5%BE%97%E8%87%AA%E5%B7%B1%E7%9A%84github-user%E4%BB%A3%E5%B8%81)
- [github实现一体化](#%E4%B8%8Egithub%E5%AE%9E%E7%8E%B0%E4%B8%80%E4%BD%93%E5%8C%96)
- [Genesis是如何运行的](#genesis%E6%98%AF%E5%A6%82%E4%BD%95%E8%BF%90%E8%A1%8C%E7%9A%84)
- [快速起步](#%E5%BF%AB%E9%80%9F%E8%B5%B7%E6%AD%A5)
- [计划](#%E8%AE%A1%E5%88%92)
- [是如何运行的](#%E5%8F%82%E4%B8%8E%E8%AE%BE%E8%AE%A1)
- [快速起步 计划](#%E6%96%87%E4%BB%B6)
- [参与设计 文件](#%E7%89%88%E6%9C%AC%E7%AE%A1%E7%90%86)
- [版本管理 设计人员](#%E8%AE%BE%E8%AE%A1%E4%BA%BA%E5%91%98)
- [许可证](#%E8%AE%B8%E5%8F%AF%E8%AF%81)

## 引言
Genesis - opensource 区块链平台，其基础工作由程序员奥列格•斯特列雷科在2011年完成的。平台源代码从零写入。参与这一项目的团
队由超过15名高级程序员组成。我们无法对我们喜爱的Genesis版本实现ICO
，为了借助Genesis发展成为世界上最优秀的区块链平台，我们决定将85%的代币免费分发给大部分程序员。

## Genesis的特点？
- 您可在Genesis 中创建带有特定规则的独立区块链生态系统，建立独立Ethereum
，它可与您网上邻居的Ethereum（Genesis的生态系统）协作。
- 在开始使用Simvolio和Protypo语言进行设计，需花4个小时掌握它们。
- 可将Simvolio和Protypo的个人设计立即装载到IOS版或Android版的手机中。为此，可使用我们即将在Appstore
和Google Play上线的软件，或者在对我们的移动应用程序源文件进行小小的更改后上传自己的设计版本。
- 平台中的全部系统数据以及统一算法都可被调整，可借助语音短信或其他算法进行更改。


## 获得自己的github-User代币！
为了预防再Genesis以及其他公共区块链平台中受到攻击，可使用GEN代币支付网络资源使用费用。将在平台
的创世块中发行1亿枚代币，其中85%（8500万GEN）将在账号存在1年以上（防止机器人）的85万GithubUser之间进行分配。我们决定选择这种方式分配代币，因为Github-User超过2400万人，他们几乎都是程序员。
<br>
为了使征集85万Github-User的过程更为有效，我们向第一批加入者提供奖励：:<br>

- 前 1,000名奖励150GEN
- 前 10,000名-140GEN
- 前 100,000名 -120GEN
- 前 500,000名- 80GEN
- 其余奖励 70GEN

<br><p align="center"><b>只需在该软件储藏仓上打星号，在平台启动时即可获得自己的代币</b>
</p><br>

[date to be announced] 征集星号活动将结束。在次日起千您需要将 [公共密钥ECDSA](http://GenesisCommunity.github.io/newkey), 上传至个人bio中
，我们将该密钥进行配对并写入Genesis-区块链中，通过密钥可访问GEN代币。如果在代币分配后还有剩余，则Genesis
储备将自行决定剩余代币的分配。
<br> <br>

如果在bio中指出Github-user用户名（在密钥空格后），您可额外获得超过10个GEN代币，向您介绍
Genesis的人同样可额外获得超过10个GEN代币（如果他在自己的设定档中上传 [ECDSA公共密钥](http://GenesisCommunity.github.io/newkey)).<br> <br>

<p align="center"><a href="https://www.facebook.com/sharer/sharer.php?u=https://github.com/GenesisCommunity/go-genesis/" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/facebook.png" width=40></a> <a href="https://twitter.com/intent/tweet?url=https%3A%2F%2Fgithub.com%2FGenesisCommunity%2Fgo-genesis&text=85%25%20of%20all%20tokens%20will%20be%20distributed%20for%20free%20among%20850,000%20GitHub%20users,%20who%20put%20a%20star%20in%20this%20repository&hashtags=genesisblockchain" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/twitter.png" width=40></a> <a href="http://reddit.com/submit?url=https://github.com/GenesisCommunity/go-genesis/&amp;title=85%25%20of%20all%20tokens%20will%20be%20distributed%20for%20free%20among%20850,000%20GitHub%20users,%20who%20put%20a%20star%20in%20this%20repository" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/reddit.png" width=40></a> <a href="https://plus.google.com/share?url=https://github.com/GenesisCommunity/go-genesis/" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/google.png" width=40></a> <a href="mailto:?subject=I wanted you to see this site&amp;body=85%25%20of%20all%20tokens%20will%20be%20distributed%20for%20free%20among%20850,000%20GitHub%20users,%20who%20put%20a%20star%20in%20this%20repository -  https://github.com/GenesisCommunity/go-genesis/" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/email.png" width=40></a> <a href="http://www.linkedin.com/shareArticle?mini=true&amp;url=https://github.com/GenesisCommunity/go-genesis/" target="_blank"><img src="https://simplesharebuttons.com/images/somacro/linkedin.png" width=40></a><br>Tell your friends!</p><br>

最后，约85万程序员可获得对区块链平台的全面监控，使用新规则建立新世界。<br> <br>

mainnet、GEN投放后完全可进入交易所。需要指出的是，目前没有任何一家公共区块链平台拥有设计个人智能合同和
在交易所发行的可能性，现有公共区块链平台中的代币的总价值低于 10亿美元。

Platform | Smart-contracts | Market Cap | Token price | Total Supply| Github | Source code
--- | --- | --- | --- | ---| --- | ---
Ethereum | [Documentation](http://solidity.readthedocs.io/en/develop/introduction-to-smart-contracts.html) | $90B | [$900](https://coinmarketcap.com/currencies/ethereum/)  | 100M | [ethereum](https://github.com/ethereum/go-ethereum) | original
NEO | [Documentation](http://docs.neo.org/en-us/sc/introduction.html) | $8B | [$120](https://coinmarketcap.com/currencies/neo/) | 100M| [neo-project](https://github.com/neo-project/neo) | original
EOS | [Documentation](https://github.com/EOSIO/eos/wiki/Smart-Contract) | $6B | [$9](https://coinmarketcap.com/currencies/eos/) | 900M | [EOSIO](https://github.com/EOSIO/eos) | original
Qtum | [Documentation](https://github.com/qtumproject/qtum/blob/master/doc/sparknet-guide.md) | $2B | [$29](https://coinmarketcap.com/currencies/qtum/) | 100M |  [qtumproject](https://github.com/qtumproject/qtum) | bitcoin fork

以 https://coinmarketcap.com/tokens/ 为基础


## 与github实现一体化
我们计划实现支付审核后的COMMIT并关闭GEN中的issue.也就是说，您可在自己的软件源中实现Genesis一体化并自动
为改善您产品的用户付款。<br>
为支持Genesis，请上传我们的识别证。<br>
<p align="center">

[![We accept GEN](https://img.shields.io/badge/We_accept-GEN-brightgreen.svg)](https://github.com/GenesisCommunity/go-genesis/)

</p>

```
[[![We accept GEN](https://img.shields.io/badge/We_accept-GEN-brightgreen.svg)](https://github.com/GenesisCommunity/go-genesis/)
```

## Genesis是如何运行的
设计  [Simvolio](http://genesiskernel.readthedocs.io/en/latest/introduction/script.html#simvolio-contracts-language)类似编程C语言，使用这一语言在字节码中写入并编辑合同。可实现最小数量的操作结构和内置功能。

<p align="center">
    <img src="https://i.imgur.com/qHosOsw.jpg">
</p><br>

创建  [Protypo](http://genesiskernel.readthedocs.io/en/latest/introduction/templates2.html#protypo-template-language). Protypo - 适用于前台的页面描述语言。实际上是样板处理器，它可将带有数据的功能顺序翻译为前台存储单元的树状表示。


<p align="center">
    <img src="https://i.imgur.com/CYL1b95.jpg">
</p>
<br>

为合同\界面和注册表数据代码的变更建立 [规则](https://genesiskernel.readthedocs.io/en/latest/introduction/what-is-Apla.html#access-rights-control-mechanism)

<p align="center">
    <img src="https://i.imgur.com/DkvR7MZ.jpg">
</p>
将个人区块链-软件上传至Play Market和App Store。<br>
https://github.com/GenesisKernel/genesis-reactnative<br><br><br>
<p align="center">
    <img src="https://i.imgur.com/m46Kxwc.png" alt="" width=250>
</p>


## 快速起步
<p align="center">
    <img src="https://i.imgur.com/6oYykyk.jpg">
</p>

https://github.com/GenesisKernel/quick-start<Br>展开

macos装配台:
```bash
bash manage.sh install 3 (抬升3个局域NOD)
```
展开linux装配台：
```bash
bash manage.sh install 3 (抬升3个局域NOD)
```
展开windows装配台:<br>
https://github.com/GenesisKernel/quick-start-win/releases<br>
```bash
win_install.exe
```


#### 控制台中的Blockexplorer
```bash
bash manage.sh db-shell 1
```
```bash
select id, time, node_position, key_id, tx from block_chain ORDER BY ID DESC LIMIT 20;
```
生成字块的NOD表单：<br>
```bash
select value from system_parameters where name='full_nodes';
```
很快可以访问 Blockexplorer 的web-版本。<br>


## 计划

我们认为，可以将源代码做的更好，因此将不断提高它的质量和产量。

#### testnet

大约在2018年3月1日投放第三个测试版本
使用个人密钥，在测试网站中对系统的工作性能进行检测。<br>

#### mainnet

大约在2018年4月15日投放<br>

## 参与设计
请阅读 [CONTRIBUTING.md](https://github.com/GenesisKernel/go-genesis/blob/master/CONTRIBUTING.md) 以获取有关Pull Requests发送过程详细信息。

## 文件
请了解并补充我们的 [文件](https://genesiskernel.readthedocs.io/)


## 版本管理
我们使用 [SemVer](http://semver.org/) 实现版本管理现有版本请见 [tags on this repository](https://github.com/GenesisKernel/go-genesis/tags)


## 设计人员

- Oleg Strelenko - founder, Initial work - https://github.com/c-darwin
- Alexey Krivonogov - core developer - https://github.com/gentee
- Alexander Boldachev - Simvolio/Protypo architecture - https://github.com/AleDvin
- Roman Potekhin - backend developer - https://github.com/potehinre
- Evgeny Lerner - backend developer - https://github.com/dvork1ng
- Dmitrij Galitskij - backend developer - https://github.com/yddmat
- Dmitriy Chertkov - backend developer - https://github.com/dchertkov
- Roman Poletaev - backend developer - https://github.com/rpoletaev
- Igor Chertov - frontend developer - https://github.com/Saurer
- Alexey Voskresenskiy - Protypo constructor developer - https://github.com/av-alex
- Vladimir Matsola - mobile developer - https://github.com/2vm
- Alex Stern - bash/python developer - https://github.com/blitzstern5
- Vasily Starovetskiy - Simvolio/Protypo developer - https://github.com/syypoo
- Andrey Voronkov - Simvolio/Protypo developer - https://github.com/CynepHy6
- Viktor Waise - Simvolio/Protypo developer - https://github.com/Waisevi
- Aleksey Sukhanov - Simvolio/Protypo developer - https://github.com/pekanius
- Yuriy Lomakin - MVP frontend, tester - https://github.com/ylomakin
- Elena Konkina - tester - https://github.com/lfreze

请查看该项目的 [参与者](https://github.com/GenesisKernel/go-genesis/graphs/contributors) 名单。<br>
[Join](mailto:hello@apla.io) Genesis团队！


## 许可证

该项目获得GPLv3许可-详情请查看文件 [LICENSE.md](https://github.com/GenesisKernel/go-genesis/blob/master/LICENSE)

<p align="center">
<a href="#"><img src="http://www.kgsbo.com/wp-content/themes/kgsbo/images/top.png" width=100 align="center"></a><br>
  <a href="#">向上</a>
</p>