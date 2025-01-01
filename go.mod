module github.com/go-sigma/sigma

go 1.23.4

require (
	code.gitea.io/sdk/gitea v0.20.0
	github.com/BurntSushi/toml v1.4.0
	github.com/IBM/sarama v1.45.0
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/alicebob/miniredis/v2 v2.34.0
	github.com/aliyun/aliyun-oss-go-sdk v3.0.2+incompatible
	github.com/anchore/syft v1.18.1
	github.com/aquasecurity/trivy v0.58.1
	github.com/avast/retry-go/v4 v4.6.0
	github.com/aws/aws-sdk-go-v2 v1.32.8
	github.com/aws/aws-sdk-go-v2/config v1.28.10
	github.com/aws/aws-sdk-go-v2/service/s3 v1.72.2
	github.com/caarlos0/env/v9 v9.0.0
	github.com/casbin/casbin/v2 v2.102.0
	github.com/casbin/gorm-adapter/v3 v3.32.0
	github.com/containers/podman/v5 v5.3.1
	github.com/deckarep/golang-set/v2 v2.7.0
	github.com/dgraph-io/badger/v4 v4.5.0
	github.com/distribution/distribution/v3 v3.0.0-rc.2
	github.com/distribution/reference v0.6.0
	github.com/docker/cli v27.4.1+incompatible
	github.com/docker/docker v27.4.1+incompatible
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.18.0
	github.com/glebarez/sqlite v1.11.0
	github.com/go-git/go-git/v5 v5.13.1
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-resty/resty/v2 v2.16.3
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang-migrate/migrate/v4 v4.18.1
	github.com/google/go-github/v53 v53.2.0
	github.com/google/uuid v1.6.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/hibiken/asynq v0.25.1
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jackc/pgx/v5 v5.7.2
	github.com/jinzhu/copier v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/labstack/echo-contrib v0.17.1
	github.com/labstack/echo/v4 v4.13.3
	github.com/matoous/go-nanoid v1.5.1
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/mholt/archiver/v3 v3.5.1
	github.com/moby/buildkit v0.18.2
	github.com/opencontainers/distribution-spec/specs-go v0.0.0-20241220184801-37e5199a9ae0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/redis/go-redis/v9 v9.7.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/xid v1.6.0
	github.com/rs/zerolog v1.33.0
	github.com/sigstore/cosign/v2 v2.4.1
	github.com/smartystreets/goconvey v1.8.1
	github.com/spf13/cast v1.7.1
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/echo-swagger v1.4.2-0.20241021205945-c9c2b320fd95
	github.com/swaggo/swag v1.16.4
	github.com/tencentyun/cos-go-sdk-v5 v0.7.60
	github.com/tidwall/gjson v1.18.0
	github.com/wagslane/go-password-validator v0.3.0
	github.com/xo/dburl v0.23.2
	gitlab.com/gitlab-org/api/client-go v0.119.0
	go.uber.org/dig v1.18.0
	go.uber.org/mock v0.5.0
	golang.org/x/crypto v0.32.0
	golang.org/x/net v0.34.0
	golang.org/x/oauth2 v0.25.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.11
	gorm.io/gen v0.3.26
	gorm.io/gorm v1.25.12
	gorm.io/plugin/dbresolver v1.5.3
	gorm.io/plugin/soft_delete v1.2.1
	k8s.io/api v0.32.0
	k8s.io/apimachinery v0.32.0
	k8s.io/client-go v0.32.0
	modernc.org/sqlite v1.34.4
)

require (
	cel.dev/expr v0.16.1 // indirect
	cloud.google.com/go v0.115.1 // indirect
	cloud.google.com/go/auth v0.9.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.5.1 // indirect
	cloud.google.com/go/iam v1.2.1 // indirect
	cloud.google.com/go/monitoring v1.21.0 // indirect
	cloud.google.com/go/storage v1.45.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/42wim/httpsig v1.2.1 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.24.1 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.48.1 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.48.1 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.12.9 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/ProtonMail/go-crypto v1.1.4 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/acobaugh/osrelease v0.1.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.0 // indirect
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/alicebob/gopher-json v0.0.0-20230218143504-906a9b012302 // indirect
	github.com/anchore/go-logger v0.0.0-20230725134548-c21dafa1ec5a // indirect
	github.com/anchore/packageurl-go v0.1.1-0.20241018175412-5c22e6360c4f // indirect
	github.com/anchore/stereoscope v0.0.11 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aquasecurity/go-version v0.0.0-20241105054539-1951e80d786f // indirect
	github.com/aquasecurity/trivy-checks v1.4.0 // indirect
	github.com/aquasecurity/trivy-db v0.0.0-20241209111357-8c398f13db0e // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.51 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.6 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/becheran/wildmatch-go v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.7.1 // indirect
	github.com/casbin/govaluate v1.2.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/cloudflare/circl v1.3.8 // indirect
	github.com/cncf/xds/go v0.0.0-20240905190251-b4127c9b8d78 // indirect
	github.com/containerd/cgroups/v3 v3.0.3 // indirect
	github.com/containerd/containerd v1.7.24 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.15.1 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/containers/buildah v1.38.0 // indirect
	github.com/containers/common v0.61.0 // indirect
	github.com/containers/image/v5 v5.33.0 // indirect
	github.com/containers/libtrust v0.0.0-20230121012942-c1716e8a8d01 // indirect
	github.com/containers/ocicrypt v1.2.0 // indirect
	github.com/containers/psgo v1.9.0 // indirect
	github.com/containers/storage v1.56.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.1-0.20231103132048-7d375ecc2b09 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20231217050601-ba74d44ecf5f // indirect
	github.com/cyphar/filepath-securejoin v0.3.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/davidmz/go-pageant v1.0.2 // indirect
	github.com/dgraph-io/ristretto v0.1.2-0.20240116140435-c67e07994f91 // indirect
	github.com/dgraph-io/ristretto/v2 v2.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitorus/pkcs7 v0.0.0-20230818184609-3a137a874352 // indirect
	github.com/digitorus/timestamp v0.0.0-20231217203849-220c5c2851b7 // indirect
	github.com/disiqueira/gotree/v3 v3.0.2 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/envoyproxy/go-control-plane v0.13.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/facebookincubator/nvdtools v0.1.5 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/github/go-spdx/v2 v2.3.2 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-fed/httpsig v1.1.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.1 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.28.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/strfmt v0.23.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.1-0.20240921181615-a817f3cc4a9e // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/certificate-transparency-go v1.2.1 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/gnostic-models v0.6.9-0.20230804172637-c7be7c783f49 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-containerregistry v0.20.2 // indirect
	github.com/google/go-intervals v0.0.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.7.6 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/hcl/v2 v2.23.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/in-toto/in-toto-golang v0.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20230811132847-661be99b8267 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/letsencrypt/boulder v0.0.0-20240620165639-de9c06129bec // indirect
	github.com/liamg/iamgo v0.0.9 // indirect
	github.com/liamg/jfather v0.0.7 // indirect
	github.com/liamg/memoryfs v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/masahiro331/go-disk v0.0.0-20240625071113-56c933208fee // indirect
	github.com/masahiro331/go-ext4-filesystem v0.0.0-20240620024024-ca14e6327bbd // indirect
	github.com/masahiro331/go-xfs-filesystem v0.0.0-20231205045356-1b22259a6c44 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/microsoft/go-mssqldb v1.7.2 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mistifyio/go-zfs/v3 v3.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/capability v0.3.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mozillazg/go-httpheader v0.4.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/open-policy-agent/opa v0.70.0 // indirect
	github.com/opencontainers/runc v1.2.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opencontainers/runtime-tools v0.9.1-0.20241001195557-6c9570a1678f // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/ostreedev/ostree-go v0.0.0-20210805093236-719684c64e4f // indirect
	github.com/owenrumney/squealer v1.2.5 // indirect
	github.com/package-url/packageurl-go v0.1.3 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.7 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/proglottis/gpgme v0.1.3 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rubenv/sql-migrate v1.7.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/sassoftware/relic v7.2.1+incompatible // indirect
	github.com/scylladb/go-set v1.0.3-0.20200225121959-cc7b2070d91e // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sigstore/fulcio v1.6.4 // indirect
	github.com/sigstore/protobuf-specs v0.3.2 // indirect
	github.com/sigstore/rekor v1.3.6 // indirect
	github.com/sigstore/sigstore v1.8.10 // indirect
	github.com/sigstore/timestamp-authority v1.2.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20230803200340-78284954bff6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/swaggo/files/v2 v2.0.1 // indirect
	github.com/sylabs/sif/v2 v2.20.1 // indirect
	github.com/sylabs/squashfs v1.0.4 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/theupdateframework/go-tuf v0.7.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/tonistiigi/go-csvvalue v0.0.0-20240710180619-ddb21b71c0b4 // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/vbatts/tar-split v0.11.6 // indirect
	github.com/vbauerster/mpb/v8 v8.8.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/wagoodman/go-partybus v0.0.0-20230516145632-8ccac152c651 // indirect
	github.com/wagoodman/go-progress v0.0.0-20230925121702-07e42b3cdba0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xanzy/go-gitlab v0.115.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/zclconf/go-cty v1.15.0 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20210826202110-33d05740a352 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.29.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.57.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/sdk v1.32.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20241210194714-1829a127f884 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/api v0.197.0 // indirect
	google.golang.org/genproto v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	google.golang.org/grpc/stats/opentelemetry v0.0.0-20240907200651-3ffb98b2c93a // indirect
	google.golang.org/protobuf v1.36.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/datatypes v1.2.5 // indirect
	gorm.io/driver/sqlite v1.5.5 // indirect
	gorm.io/driver/sqlserver v1.5.4 // indirect
	gorm.io/hints v1.1.2 // indirect
	helm.sh/helm/v3 v3.16.3 // indirect
	k8s.io/apiextensions-apiserver v0.31.1 // indirect
	k8s.io/apiserver v0.31.2 // indirect
	k8s.io/cli-runtime v0.31.2 // indirect
	k8s.io/component-base v0.31.2 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/kubectl v0.31.2 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	modernc.org/gc/v3 v3.0.0-20240801135723-a856999a2e4a // indirect
	modernc.org/libc v1.59.9 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
	mvdan.cc/sh/v3 v3.10.0 // indirect
	oras.land/oras-go v1.2.5 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/kustomize/api v0.17.2 // indirect
	sigs.k8s.io/kustomize/kyaml v0.17.1 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
	tags.cncf.io/container-device-interface v0.8.0 // indirect
)

replace gorm.io/plugin/soft_delete => github.com/go-sigma/soft_delete v0.0.0-20240610034828-c89bcecb7e29

replace github.com/glebarez/go-sqlite => github.com/go-sigma/go-sqlite v0.0.0-20240610040849-04f3543f8179

replace github.com/glebarez/sqlite => github.com/go-sigma/sqlite v0.0.0-20240610040657-11b421fdf449
