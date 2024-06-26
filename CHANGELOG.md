
<a name="v1.3.0"></a>
## [v1.3.0](https://github.com/go-sigma/sigma/compare/v1.3.0-rc.1...v1.3.0) (2024-06-11)

### Features

* Push image to tcr ([#371](https://github.com/go-sigma/sigma/issues/371))
* Push image to acr ([#370](https://github.com/go-sigma/sigma/issues/370))
* Upgrade aws s3 v2 ([#369](https://github.com/go-sigma/sigma/issues/369))
* Login on enter key down ([#367](https://github.com/go-sigma/sigma/issues/367))
* Add icon on login page ([#366](https://github.com/go-sigma/sigma/issues/366))

### Docs

* Update changelog ([#365](https://github.com/go-sigma/sigma/issues/365))


<a name="v1.3.0-rc.1"></a>
## [v1.3.0-rc.1](https://github.com/go-sigma/sigma/compare/v1.2.0...v1.3.0-rc.1) (2024-05-05)

### Features

* Update config ([#364](https://github.com/go-sigma/sigma/issues/364))
* Reimplement cache with badger ([#362](https://github.com/go-sigma/sigma/issues/362))
* Update UploadPart impl ([#361](https://github.com/go-sigma/sigma/issues/361))
* No limit on the number of ns that the admin user can create ([#360](https://github.com/go-sigma/sigma/issues/360))
* Change the locker initialize ([#359](https://github.com/go-sigma/sigma/issues/359))
* Reimplement lock with badger ([#358](https://github.com/go-sigma/sigma/issues/358))
* Change distribution user get func ([#356](https://github.com/go-sigma/sigma/issues/356))
* Add test params ([#354](https://github.com/go-sigma/sigma/issues/354))
* Add codeql ([#353](https://github.com/go-sigma/sigma/issues/353))
* Reimplement lock ([#352](https://github.com/go-sigma/sigma/issues/352))
* Support disbale oauth2 and anonymous login ([#350](https://github.com/go-sigma/sigma/issues/350))
* Support disable push base image after server started ([#347](https://github.com/go-sigma/sigma/issues/347))
* Add workqueue inmemory implement ([#345](https://github.com/go-sigma/sigma/issues/345))
* Update blob cacher ([#342](https://github.com/go-sigma/sigma/issues/342))
* Copy dockerfile and builder image after server started ([#341](https://github.com/go-sigma/sigma/issues/341))
* Add builder and dockerfile to sigma ([#340](https://github.com/go-sigma/sigma/issues/340))

### Bug Fixes

* Fix workq concurrency map ([#363](https://github.com/go-sigma/sigma/issues/363))
* Fix gc ([#349](https://github.com/go-sigma/sigma/issues/349))

### Docs

* Add zh readme ([#338](https://github.com/go-sigma/sigma/issues/338))

### Upgrade

* Upgrade docker ([#348](https://github.com/go-sigma/sigma/issues/348))
* Upgrade podman ([#346](https://github.com/go-sigma/sigma/issues/346))

### Unit Tests

* Update unit test ([#357](https://github.com/go-sigma/sigma/issues/357))
* Update test ci config ([#355](https://github.com/go-sigma/sigma/issues/355))
* Add unit test for blob gc ([#351](https://github.com/go-sigma/sigma/issues/351))
* Update dao test ([#344](https://github.com/go-sigma/sigma/issues/344))
* Add unit test for gc artifact ([#339](https://github.com/go-sigma/sigma/issues/339))


<a name="v1.2.0"></a>
## [v1.2.0](https://github.com/go-sigma/sigma/compare/v1.2.0-rc.1...v1.2.0) (2024-03-11)

### Features

* Rename artifact_artifacts sub artifact field ([#337](https://github.com/go-sigma/sigma/issues/337))

### Bug Fixes

* Fix gc artifact for multi-arch manifest ([#336](https://github.com/go-sigma/sigma/issues/336))


<a name="v1.2.0-rc.1"></a>
## [v1.2.0-rc.1](https://github.com/go-sigma/sigma/compare/v1.1.0...v1.2.0-rc.1) (2024-03-11)

### Features

* Change time field type to int64 ([#332](https://github.com/go-sigma/sigma/issues/332))
* Add log field for sql caller ([#329](https://github.com/go-sigma/sigma/issues/329))
* Support gc trigger webhook ([#323](https://github.com/go-sigma/sigma/issues/323))
* Update trivy command ([#320](https://github.com/go-sigma/sigma/issues/320))
* Support webhook in admin ([#319](https://github.com/go-sigma/sigma/issues/319))
* Skip etag for the api ([#318](https://github.com/go-sigma/sigma/issues/318))
* Change pprof path ([#316](https://github.com/go-sigma/sigma/issues/316))
* Update pprof config ([#315](https://github.com/go-sigma/sigma/issues/315))
* Add etag ([#313](https://github.com/go-sigma/sigma/issues/313))
* Add metrics ([#312](https://github.com/go-sigma/sigma/issues/312))
* Move builder helper func ([#308](https://github.com/go-sigma/sigma/issues/308))
* Always check remote proxy server ([#307](https://github.com/go-sigma/sigma/issues/307))
* Add webhook log detail info ([#304](https://github.com/go-sigma/sigma/issues/304))
* Support webhook ([#303](https://github.com/go-sigma/sigma/issues/303))
* Support delete tag on page ([#301](https://github.com/go-sigma/sigma/issues/301))
* Check permission in pages ([#300](https://github.com/go-sigma/sigma/issues/300))
* Add role field for namespace api ([#299](https://github.com/go-sigma/sigma/issues/299))
* Add support oss ([#298](https://github.com/go-sigma/sigma/issues/298))
* Add podman for builder ([#297](https://github.com/go-sigma/sigma/issues/297))

### Bug Fixes

* Fix gc ([#334](https://github.com/go-sigma/sigma/issues/334))
* Fix login with incorrect username or password ([#330](https://github.com/go-sigma/sigma/issues/330))
* Fix audit resource_type field ([#328](https://github.com/go-sigma/sigma/issues/328))
* Fix member policy field ([#326](https://github.com/go-sigma/sigma/issues/326))
* Fix user create or update failed ([#325](https://github.com/go-sigma/sigma/issues/325))
* Fix list namespace auth ([#314](https://github.com/go-sigma/sigma/issues/314))

### Docs

* Add wechat ([#327](https://github.com/go-sigma/sigma/issues/327))
* Update readme ([#317](https://github.com/go-sigma/sigma/issues/317))
* Update readme ([#296](https://github.com/go-sigma/sigma/issues/296))

### Upgrade

* Upgrade docker and minio ([#295](https://github.com/go-sigma/sigma/issues/295))
* Upgrade runc for avoid cve ([#294](https://github.com/go-sigma/sigma/issues/294))

### CI

* Update ci config ([#335](https://github.com/go-sigma/sigma/issues/335))
* Add errors test case ([#322](https://github.com/go-sigma/sigma/issues/322))
* Update codecov config ([#321](https://github.com/go-sigma/sigma/issues/321))
* Add docker hub ([#311](https://github.com/go-sigma/sigma/issues/311))
* Update ci ([#302](https://github.com/go-sigma/sigma/issues/302))


<a name="v1.1.0"></a>
## [v1.1.0](https://github.com/go-sigma/sigma/compare/v1.0.0...v1.1.0) (2024-02-04)

### Features

* Support build with mirror ([#285](https://github.com/go-sigma/sigma/issues/285))
* Create namespace send webhook ([#283](https://github.com/go-sigma/sigma/issues/283))
* Support apptainer sif ([#278](https://github.com/go-sigma/sigma/issues/278))
* Implement blob redirect ([#275](https://github.com/go-sigma/sigma/issues/275))
* Support disable image builder ([#273](https://github.com/go-sigma/sigma/issues/273))
* Use zig as cross compiler ([#269](https://github.com/go-sigma/sigma/issues/269))
* Support disable image build ([#267](https://github.com/go-sigma/sigma/issues/267))

### Bug Fixes

* Fix artifact type ([#276](https://github.com/go-sigma/sigma/issues/276))
* Fix builder image not exist ([#272](https://github.com/go-sigma/sigma/issues/272))
* Enable cgo for sqlite3 ([#268](https://github.com/go-sigma/sigma/issues/268))

### Docs

* Update readme ([#292](https://github.com/go-sigma/sigma/issues/292))
* Add apptainer docs ([#282](https://github.com/go-sigma/sigma/issues/282))

### Upgrade

* Upgrade dependencies ([#290](https://github.com/go-sigma/sigma/issues/290))
* Upgrade distribution ([#277](https://github.com/go-sigma/sigma/issues/277))
* Update dependencies ([#271](https://github.com/go-sigma/sigma/issues/271))

### CI

* Update ci config ([#288](https://github.com/go-sigma/sigma/issues/288))
* Update tag release config ([#266](https://github.com/go-sigma/sigma/issues/266))
* Update tag release config ([#265](https://github.com/go-sigma/sigma/issues/265))

### Unit Tests

* Add test for auth ([#289](https://github.com/go-sigma/sigma/issues/289))
* Add namespace update ut ([#287](https://github.com/go-sigma/sigma/issues/287))
* Add list namespace ut ([#286](https://github.com/go-sigma/sigma/issues/286))
* Add delete namespace ut back ([#284](https://github.com/go-sigma/sigma/issues/284))
* Add unit test for signing init ([#281](https://github.com/go-sigma/sigma/issues/281))
* Add unit test for workq ([#280](https://github.com/go-sigma/sigma/issues/280))
* Add unit test for audit ([#279](https://github.com/go-sigma/sigma/issues/279))
* Add unit test for inmemory cache ([#270](https://github.com/go-sigma/sigma/issues/270))


<a name="v1.0.0"></a>
## v1.0.0 (2023-12-29)

### Features

* Add changlog config ([#263](https://github.com/go-sigma/sigma/issues/263))
* Add introduction for sigma ([#260](https://github.com/go-sigma/sigma/issues/260))
* Show elapsed on builder page ([#257](https://github.com/go-sigma/sigma/issues/257))
* Change cache api ([#256](https://github.com/go-sigma/sigma/issues/256))
* Support update builder show the selected code repository ([#254](https://github.com/go-sigma/sigma/issues/254))
* Support code repository image build ([#253](https://github.com/go-sigma/sigma/issues/253))
* Auto refresh token if tab was active ([#252](https://github.com/go-sigma/sigma/issues/252))
* Redirect the login if not login ([#251](https://github.com/go-sigma/sigma/issues/251))
* Support anonymous login ([#250](https://github.com/go-sigma/sigma/issues/250))
* Add overview field for namespace ([#249](https://github.com/go-sigma/sigma/issues/249))
* Remove repository visibility field ([#248](https://github.com/go-sigma/sigma/issues/248))
* Check auth in distribution manifest api ([#247](https://github.com/go-sigma/sigma/issues/247))
* Check auth in distribution api ([#246](https://github.com/go-sigma/sigma/issues/246))
* Add anonymous user ([#245](https://github.com/go-sigma/sigma/issues/245))
* Change dal test case ([#243](https://github.com/go-sigma/sigma/issues/243))
* Change repository and tag api ([#242](https://github.com/go-sigma/sigma/issues/242))
* Add docs for namespace member api ([#240](https://github.com/go-sigma/sigma/issues/240))
* Change namespaces created_at field type ([#239](https://github.com/go-sigma/sigma/issues/239))
* Upgrade yarn 4.0.2 ([#237](https://github.com/go-sigma/sigma/issues/237))
* Add auth and audit for namespac ([#236](https://github.com/go-sigma/sigma/issues/236))
* Change auth service implement ([#235](https://github.com/go-sigma/sigma/issues/235))
* Add auth service ([#234](https://github.com/go-sigma/sigma/issues/234))
* Update minio version ([#233](https://github.com/go-sigma/sigma/issues/233))
* Support anonymous list namespace api ([#232](https://github.com/go-sigma/sigma/issues/232))
* Add member page for namespace ([#231](https://github.com/go-sigma/sigma/issues/231))
* Change sample in dind ([#229](https://github.com/go-sigma/sigma/issues/229))
* Change handler variable ([#228](https://github.com/go-sigma/sigma/issues/228))
* Add run btn on gc runner page ([#227](https://github.com/go-sigma/sigma/issues/227))
* Add gc repository, tag, artifact, blob daemon task ([#226](https://github.com/go-sigma/sigma/issues/226))
* Add quick helm deploy command ([#223](https://github.com/go-sigma/sigma/issues/223))
* Remove lfs files ([#224](https://github.com/go-sigma/sigma/issues/224))
* Support image spec 1.1 referrers api ([#222](https://github.com/go-sigma/sigma/issues/222))
* Add multiarch sample images ([#221](https://github.com/go-sigma/sigma/issues/221))
* Change tags page ([#220](https://github.com/go-sigma/sigma/issues/220))
* Support create user and update user ([#219](https://github.com/go-sigma/sigma/issues/219))
* Support filter type for list tags ([#218](https://github.com/go-sigma/sigma/issues/218))
* Support sign image ([#217](https://github.com/go-sigma/sigma/issues/217))
* Support parse tag with digest ([#216](https://github.com/go-sigma/sigma/issues/216))
* Auth user in middleware ([#215](https://github.com/go-sigma/sigma/issues/215))
* Add version api ([#214](https://github.com/go-sigma/sigma/issues/214))
* Filter the cosign ([#213](https://github.com/go-sigma/sigma/issues/213))
* Support show helm and docker icon ([#212](https://github.com/go-sigma/sigma/issues/212))
* Add setting ([#211](https://github.com/go-sigma/sigma/issues/211))
* Update user profile and password ([#210](https://github.com/go-sigma/sigma/issues/210))
* Design docs start page ([#209](https://github.com/go-sigma/sigma/issues/209))
* Remove redis code ([#207](https://github.com/go-sigma/sigma/issues/207))
* Add sample script for demo server ([#206](https://github.com/go-sigma/sigma/issues/206))
* Set network for builder ([#205](https://github.com/go-sigma/sigma/issues/205))
* Remove viper from storage ([#203](https://github.com/go-sigma/sigma/issues/203))
* Add friendly link ([#202](https://github.com/go-sigma/sigma/issues/202))
* Add sigma builder to image ([#201](https://github.com/go-sigma/sigma/issues/201))
* Add demo server ([#200](https://github.com/go-sigma/sigma/issues/200))
* Remove redis ([#198](https://github.com/go-sigma/sigma/issues/198))
* Add option for workq producer ([#197](https://github.com/go-sigma/sigma/issues/197))
* Add workq redis implemention ([#196](https://github.com/go-sigma/sigma/issues/196))
* Change cacher and locker implemention ([#195](https://github.com/go-sigma/sigma/issues/195))
* Support workq ([#194](https://github.com/go-sigma/sigma/issues/194))
* Support update builder ([#193](https://github.com/go-sigma/sigma/issues/193))
* Support create and update builder ([#192](https://github.com/go-sigma/sigma/issues/192))
* Complete builder setup ([#191](https://github.com/go-sigma/sigma/issues/191))
* Change builder table definition ([#190](https://github.com/go-sigma/sigma/issues/190))
* Builder setup page ([#189](https://github.com/go-sigma/sigma/issues/189))
* Support code repository resync ([#188](https://github.com/go-sigma/sigma/issues/188))
* Support user signed grant oauth2 provider ([#187](https://github.com/go-sigma/sigma/issues/187))
* Add code repo owners is_org field ([#186](https://github.com/go-sigma/sigma/issues/186))
* Support cos ([#185](https://github.com/go-sigma/sigma/issues/185))
* Add code repository page ([#184](https://github.com/go-sigma/sigma/issues/184))
* Display providers ([#183](https://github.com/go-sigma/sigma/issues/183))
* Add handler for setup builder ([#181](https://github.com/go-sigma/sigma/issues/181))
* Add cronjob for builder ([#180](https://github.com/go-sigma/sigma/issues/180))
* Add handler for code repository ([#179](https://github.com/go-sigma/sigma/issues/179))
* Code repository ([#178](https://github.com/go-sigma/sigma/issues/178))
* Add graceful ([#176](https://github.com/go-sigma/sigma/issues/176))
* Add k8s builder implemention ([#175](https://github.com/go-sigma/sigma/issues/175))
* Delete config file before create in builder ([#174](https://github.com/go-sigma/sigma/issues/174))
* Change default password ([#173](https://github.com/go-sigma/sigma/issues/173))
* Support image builder ([#172](https://github.com/go-sigma/sigma/issues/172))
* Add builder dao query funcs ([#170](https://github.com/go-sigma/sigma/issues/170))
* Add crypt utils ([#169](https://github.com/go-sigma/sigma/issues/169))
* Add builder and builder logs table ([#167](https://github.com/go-sigma/sigma/issues/167))
* Add func for strings join ([#168](https://github.com/go-sigma/sigma/issues/168))
* Add func for split dir with slash ([#165](https://github.com/go-sigma/sigma/issues/165))
* Support builder with buildkit ([#164](https://github.com/go-sigma/sigma/issues/164))
* Add webhook ([#163](https://github.com/go-sigma/sigma/issues/163))
* Add daemon task log table ([#162](https://github.com/go-sigma/sigma/issues/162))
* Add config struct ([#161](https://github.com/go-sigma/sigma/issues/161))
* Support reset password ([#160](https://github.com/go-sigma/sigma/issues/160))
* Support hot namespace and logout ([#158](https://github.com/go-sigma/sigma/issues/158))
* Auto create namespace ([#155](https://github.com/go-sigma/sigma/issues/155))
* Support GitHub oauth2 login ([#154](https://github.com/go-sigma/sigma/issues/154))
* Add redis in dockerfile ([#153](https://github.com/go-sigma/sigma/issues/153))
* Update helm chart ([#152](https://github.com/go-sigma/sigma/issues/152))
* List tag with type ([#149](https://github.com/go-sigma/sigma/issues/149))
* Add bytemd support edit repo summary ([#148](https://github.com/go-sigma/sigma/issues/148))
* Add server domain config ([#147](https://github.com/go-sigma/sigma/issues/147))
* Display linux distro ([#142](https://github.com/go-sigma/sigma/issues/142))
* Tag page just one request ([#141](https://github.com/go-sigma/sigma/issues/141))
* Support tags page ([#140](https://github.com/go-sigma/sigma/issues/140))
* Support repository page ([#137](https://github.com/go-sigma/sigma/issues/137))
* Replace limit,last with limit,offset ([#135](https://github.com/go-sigma/sigma/issues/135))
* Fix toast close automatic ([#134](https://github.com/go-sigma/sigma/issues/134))
* Support update and delete namespace ([#133](https://github.com/go-sigma/sigma/issues/133))
* List namespace sort with field ([#132](https://github.com/go-sigma/sigma/issues/132))
* Add new sort header component ([#131](https://github.com/go-sigma/sigma/issues/131))
* Implement quota component ([#130](https://github.com/go-sigma/sigma/issues/130))
* Support tag limit ([#129](https://github.com/go-sigma/sigma/issues/129))
* Support namespace list ([#127](https://github.com/go-sigma/sigma/issues/127))
* Complete user login ([#126](https://github.com/go-sigma/sigma/issues/126))
* Support all in one ([#125](https://github.com/go-sigma/sigma/issues/125))
* Add logo for project ([#124](https://github.com/go-sigma/sigma/issues/124))
* Add repository create and update api ([#123](https://github.com/go-sigma/sigma/issues/123))
* Add description for repository ([#122](https://github.com/go-sigma/sigma/issues/122))
* Support github oauth2 ([#121](https://github.com/go-sigma/sigma/issues/121))
* Move pages to web-next ([#120](https://github.com/go-sigma/sigma/issues/120))
* Add cacher utils ([#119](https://github.com/go-sigma/sigma/issues/119))
* Implement gc ([#117](https://github.com/go-sigma/sigma/issues/117))
* Add custom func for casbin validate ([#116](https://github.com/go-sigma/sigma/issues/116))
* Add casbin rules for validate policy ([#115](https://github.com/go-sigma/sigma/issues/115))
* Add visibility for namespace ([#113](https://github.com/go-sigma/sigma/issues/113))
* Add namespace quota ([#112](https://github.com/go-sigma/sigma/issues/112))
* Convert all fields corresponding to numbers in the database to int64 ([#107](https://github.com/go-sigma/sigma/issues/107))
* Add transaction for put manifest ([#101](https://github.com/go-sigma/sigma/issues/101))
* Support standard distribution-spec api ([#94](https://github.com/go-sigma/sigma/issues/94))
* User sonic as default json serializer ([#93](https://github.com/go-sigma/sigma/issues/93))
* Version command never need config file ([#92](https://github.com/go-sigma/sigma/issues/92))
* Namespace associate with user ([#90](https://github.com/go-sigma/sigma/issues/90))
* Add proxy task tag ([#81](https://github.com/go-sigma/sigma/issues/81))
* Check filesystem api ([#79](https://github.com/go-sigma/sigma/issues/79))
* Support proxy artifact ([#78](https://github.com/go-sigma/sigma/issues/78))
* Add clients api ([#76](https://github.com/go-sigma/sigma/issues/76))
* Change logger ctx ([#75](https://github.com/go-sigma/sigma/issues/75))
* Separate token router ([#71](https://github.com/go-sigma/sigma/issues/71))
* Add helm one-click deployment ([#67](https://github.com/go-sigma/sigma/issues/67))
* Move enums to one package ([#59](https://github.com/go-sigma/sigma/issues/59))
* Support proxy blobs ([#53](https://github.com/go-sigma/sigma/issues/53))
* Support proxy manifest ([#51](https://github.com/go-sigma/sigma/issues/51))
* Add distribution clients ([#50](https://github.com/go-sigma/sigma/issues/50))
* Change gorm logger to zero log ([#48](https://github.com/go-sigma/sigma/issues/48))
* Add nextjs as new fe ([#47](https://github.com/go-sigma/sigma/issues/47))
* Support scan and sbom ([#46](https://github.com/go-sigma/sigma/issues/46))
* Save sbom and scan ([#41](https://github.com/go-sigma/sigma/issues/41))
* Add helm chart ([#40](https://github.com/go-sigma/sigma/issues/40))
* Complete sbom and scan worker ([#39](https://github.com/go-sigma/sigma/issues/39))
* Support basic auth ([#37](https://github.com/go-sigma/sigma/issues/37))
* Update README.md ([#33](https://github.com/go-sigma/sigma/issues/33))
* Support user login, logout and docker login ([#28](https://github.com/go-sigma/sigma/issues/28))
* Add password generator and verify ([#27](https://github.com/go-sigma/sigma/issues/27))
* Add config checker ([#26](https://github.com/go-sigma/sigma/issues/26))
* Add test for server handler ([#22](https://github.com/go-sigma/sigma/issues/22))
* Add counter test case
* Implement simple time wheel
* Support multi arch image
* Implement token gen and revoke
* Implement leader election with redis dist lock
* Implement user login
* Adjust pull times
* Add swagger
* Change int64 to uint64
* Implement ns and repo fe
* Add artifact and tag api
* Add artifact and tag api
* Implement leader election with k8s
* Implement notification
* Add errcode and response data definition
* Implement filesystem
* Support create namespace on page

### Bug Fixes

* Fix docs build ([#261](https://github.com/go-sigma/sigma/issues/261))
* Fix cosign ([#259](https://github.com/go-sigma/sigma/issues/259))
* Fix distribution router ([#83](https://github.com/go-sigma/sigma/issues/83))
* Init config in server and worker ([#52](https://github.com/go-sigma/sigma/issues/52))
* Fix HTTP Range header byte ranges are inclusive ([#45](https://github.com/go-sigma/sigma/issues/45))
* Fix docker login ([#38](https://github.com/go-sigma/sigma/issues/38))
* Apply fixes from CodeFactor ([#21](https://github.com/go-sigma/sigma/issues/21))

### Code Refactoring

* Refactor put manifest  ([#99](https://github.com/go-sigma/sigma/issues/99))
* Refactor dao func ([#98](https://github.com/go-sigma/sigma/issues/98))
* Refactor artifact dao query ([#97](https://github.com/go-sigma/sigma/issues/97))
* Refactor manifest router ([#95](https://github.com/go-sigma/sigma/issues/95))

### Docs

* Add more swagger api docs ([#49](https://github.com/go-sigma/sigma/issues/49))
* Add badge ([#30](https://github.com/go-sigma/sigma/issues/30))
* Add README description ([#16](https://github.com/go-sigma/sigma/issues/16))

### Upgrade

* Upgrade go version 1.20.7 ([#182](https://github.com/go-sigma/sigma/issues/182))
* Upgrade dependencies ([#171](https://github.com/go-sigma/sigma/issues/171))
* Upgrade dependencies ([#100](https://github.com/go-sigma/sigma/issues/100))
* Bump vite from 4.2.1 to 4.2.3 in /web ([#88](https://github.com/go-sigma/sigma/issues/88))
* Update deps ([#36](https://github.com/go-sigma/sigma/issues/36))
* Bump github.com/containerd/containerd from 1.6.1 to 1.6.18 ([#25](https://github.com/go-sigma/sigma/issues/25))

### CI

* Disable ci build with timestamp ([#241](https://github.com/go-sigma/sigma/issues/241))
* Run pr without secret ([#225](https://github.com/go-sigma/sigma/issues/225))
* Branch main will not push image ([#157](https://github.com/go-sigma/sigma/issues/157))
* Add debian based image ([#156](https://github.com/go-sigma/sigma/issues/156))
* Use go build cache ([#96](https://github.com/go-sigma/sigma/issues/96))
* Update CI ([#91](https://github.com/go-sigma/sigma/issues/91))
* Never push image in normal ci ([#66](https://github.com/go-sigma/sigma/issues/66))
* Add github ci actions

### Unit Tests

* Change distribution test case ([#244](https://github.com/go-sigma/sigma/issues/244))
* Change inits test case ([#238](https://github.com/go-sigma/sigma/issues/238))
* Add test case for logout ([#159](https://github.com/go-sigma/sigma/issues/159))
* Add unit test for namespace ([#114](https://github.com/go-sigma/sigma/issues/114))
* Add unit test for xerrors ([#109](https://github.com/go-sigma/sigma/issues/109))
* Add unit test for put manifest ([#105](https://github.com/go-sigma/sigma/issues/105))
* Add e2e test ([#89](https://github.com/go-sigma/sigma/issues/89))
* Add unit test for dao ([#87](https://github.com/go-sigma/sigma/issues/87))
* Add unit test for dao ([#86](https://github.com/go-sigma/sigma/issues/86))
* Add unit test for artifact ([#84](https://github.com/go-sigma/sigma/issues/84))
* Add unit test for s3 ([#80](https://github.com/go-sigma/sigma/issues/80))
* Add unit test for repository ([#77](https://github.com/go-sigma/sigma/issues/77))
* Add unit test for namespace ([#74](https://github.com/go-sigma/sigma/issues/74))
* Add test files ([#73](https://github.com/go-sigma/sigma/issues/73))
* Add test case for dao ([#72](https://github.com/go-sigma/sigma/issues/72))
* Add unit test for daemon ([#70](https://github.com/go-sigma/sigma/issues/70))
* Add unit test for dal ([#69](https://github.com/go-sigma/sigma/issues/69))
* Add unit test for user router ([#68](https://github.com/go-sigma/sigma/issues/68))
* Add unit test for router ([#63](https://github.com/go-sigma/sigma/issues/63))
* Add unit test for xerrors ([#62](https://github.com/go-sigma/sigma/issues/62))
* Add unit test for middlewares ([#61](https://github.com/go-sigma/sigma/issues/61))
* Add unit test for proxy dao ([#60](https://github.com/go-sigma/sigma/issues/60))
* Add unit test for validator ([#58](https://github.com/go-sigma/sigma/issues/58))
* Add unit test for validator ([#57](https://github.com/go-sigma/sigma/issues/57))
* Add unit test ([#56](https://github.com/go-sigma/sigma/issues/56))
* Add limit reader unit test ([#54](https://github.com/go-sigma/sigma/issues/54))
* Add test for inits ([#44](https://github.com/go-sigma/sigma/issues/44))
* Add ut for filesystem ([#32](https://github.com/go-sigma/sigma/issues/32))
* Update token test ([#31](https://github.com/go-sigma/sigma/issues/31))
* Add unit test for s3 interface ([#24](https://github.com/go-sigma/sigma/issues/24))

