INSERT INTO "daemon_gc_artifact_rules" ("id", "namespace_id", "is_running", "retention_day", "cron_enabled", "cron_rule", "cron_next_trigger", "created_at", "updated_at", "deleted_at")
  VALUES (1, NULL, 0, 0, 0, NULL, NULL, 1710257712890, 1710257712890, 0);

INSERT INTO "daemon_gc_artifact_runners" ("id", "rule_id", "message", "status", "operate_type", "operate_user_id", "started_at", "ended_at", "duration", "success_count", "failed_count", "created_at", "updated_at", "deleted_at")
  VALUES (1, 1, NULL, 'Pending', 'Manual', NULL, NULL, NULL, NULL, NULL, NULL, 1710257941577, 1710257943729, 0);

INSERT INTO "repositories" ("id", "name", "description", "overview", "size_limit", "size", "tag_limit", "tag_count", "namespace_id", "created_at", "updated_at", "deleted_at")
  VALUES (1, 'library/alpine', NULL, NULL, 0, 13709154, 0, 1, 1, 1710256613610, 1710256613610, 0);

INSERT INTO "artifacts" ("id", "namespace_id", "repository_id", "digest", "size", "blobs_size", "content_type", "raw", "config_raw", "config_media_type", "type", "pushed_at", "last_pull", "pull_times", "referrer_id", "created_at", "updated_at", "deleted_at")
  VALUES (1, 1, 1, 'sha256:e89271a298507dd66a2f116824cf942735b3cac13cf7da045a662a3b1146e595', 480, 3343448, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a62303538646363663662653533373134623063633333346332313863613535333166663663353239613339316430643638636166303064626434316437303561222c0a202020202273697a65223a203630300a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6c617965722e76312e7461722b677a6970222c0a20202020202022646967657374223a20227368613235363a30383430396434313732363033663430623536656236623736323430613165366264373862616130653936353930646337666637366335663161303933616632222c0a2020202020202273697a65223a20333334323834380a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a2261726d3634222c22636f6e666967223a7b22456e76223a5b22504154483d2f7573722f6c6f63616c2f7362696e3a2f7573722f6c6f63616c2f62696e3a2f7573722f7362696e3a2f7573722f62696e3a2f7362696e3a2f62696e225d2c22436d64223a5b222f62696e2f7368225d2c224f6e4275696c64223a6e756c6c7d2c2263726561746564223a22323032332d30352d30395432333a31313a30382e3237373532333139365a222c22686973746f7279223a5b7b2263726561746564223a22323032332d30352d30395432333a31313a30382e3038393835393939315a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f7029204144442066696c653a6466376663636333343533623665633134303164323761313239356230383832613833653733316664653866323364623964336636383761326236623465373020696e202f20227d2c7b2263726561746564223a22323032332d30352d30395432333a31313a30382e3237373532333139365a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f70292020434d44205b5c222f62696e2f73685c225d222c22656d7074795f6c61796572223a747275657d5d2c226f73223a226c696e7578222c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a39346464376435333166613536393563306330333364636236396632313363326234633362356133616536653439373235326261383864613837313639633366225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Image', 1710256613617, 0, 0, NULL, 1710256613617, 1710256613617, 0),
  (2, 1, 1, 'sha256:f2ad7800cb3ae5ac2063a3db6edc56fc7ea975c3b83dd7cfbd1a9b0d9476e352', 480, 3398090, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a61373565333534626535613665373862653864343833376235306238663231643539366232313364323237353332396261346265336662653634343039326437222c0a202020202273697a65223a203630300a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6c617965722e76312e7461722b677a6970222c0a20202020202022646967657374223a20227368613235363a38613439666462336236613566663262643865633661383663303562323932326130663734353435373965636330373633376539346466643164303633396236222c0a2020202020202273697a65223a20333339373439300a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22616d643634222c22636f6e666967223a7b22456e76223a5b22504154483d2f7573722f6c6f63616c2f7362696e3a2f7573722f6c6f63616c2f62696e3a2f7573722f7362696e3a2f7573722f62696e3a2f7362696e3a2f62696e225d2c22436d64223a5b222f62696e2f7368225d2c224f6e4275696c64223a6e756c6c7d2c2263726561746564223a22323032332d30352d30395432333a31313a31302e3133323134373532365a222c22686973746f7279223a5b7b2263726561746564223a22323032332d30352d30395432333a31313a31302e3030373231373535335a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f7029204144442066696c653a3736323564646664353839666238323465653339663162316562333837623938663336373634323066663532663236656239643937353135316538383936363720696e202f20227d2c7b2263726561746564223a22323032332d30352d30395432333a31313a31302e3133323134373532365a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f70292020434d44205b5c222f62696e2f73685c225d222c22656d7074795f6c61796572223a747275657d5d2c226f73223a226c696e7578222c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a62623031626437653332623538623636393463386333363232633233303137316631636563323430303161383230363861386433306433333866343230643663225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Image', 1710256613639, 0, 0, NULL, 1710256613639, 1710256613639, 0),
  (3, 1, 1, 'sha256:fd790d0563e793f0f736cbcaa6d45034a2402ed8e6ed05857c9df0913d94d4d8', 838, 83080, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a30393466643934616636343965376363313066636437666438363763383839393338363662613161373130376536323736656366663364306366326334346639222c0a202020202273697a65223a203234310a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a32303664653134633033643434623233376561313738326337663266303538663631343161623133303331316238323839626561313936346438636237366230222c0a2020202020202273697a65223a2038313535352c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f737064782e6465762f446f63756d656e74220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a30363061666362623836653731353130313639333536343162323230313463643331353266636239336637383031363834656637366433666538336265336638222c0a2020202020202273697a65223a20313238342c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f736c73612e6465762f70726f76656e616e63652f76302e32220a2020202020207d0a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22756e6b6e6f776e222c226f73223a22756e6b6e6f776e222c22636f6e666967223a7b7d2c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a32303664653134633033643434623233376561313738326337663266303538663631343161623133303331316238323839626561313936346438636237366230222c227368613235363a30363061666362623836653731353130313639333536343162323230313463643331353266636239336637383031363834656637366433666538336265336638225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Provenance', 1710256613662, 0, 0, NULL, 1710256613662, 1710256613662, 0),
  (4, 1, 1, 'sha256:14097473125d074c4d51962a4d3f23a705e6bd75e8b7ce288671ba2a89aeb290', 838, 83098, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a35393664663337366466633334326436363934323738373865646333656365316464363133636261343135376235653832316166323136336134616465613631222c0a202020202273697a65223a203234310a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a36633738643666373661316565636163343838666462343736343638333262303838303463306437366230663531316537323839303162306164646437393061222c0a2020202020202273697a65223a2038313537332c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f737064782e6465762f446f63756d656e74220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a36356536633962316365613039373738653136363934336131353530663934376462363564613939333737616232373038373461373431386534326435316266222c0a2020202020202273697a65223a20313238342c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f736c73612e6465762f70726f76656e616e63652f76302e32220a2020202020207d0a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22756e6b6e6f776e222c226f73223a22756e6b6e6f776e222c22636f6e666967223a7b7d2c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a36633738643666373661316565636163343838666462343736343638333262303838303463306437366230663531316537323839303162306164646437393061222c227368613235363a36356536633962316365613039373738653136363934336131353530663934376462363564613939333737616232373038373461373431386534326435316266225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Provenance', 1710256613676, 0, 0, NULL, 1710256613676, 1710256613676, 0),
  (5, 1, 1, 'sha256:9e9b0bcd760277d309c976b6f99ca7e6171e143a954c34bd79ff7150d4633453', 1607, 2636, 'application/vnd.oci.image.index.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e696e6465782e76312b6a736f6e222c0a2020226d616e696665737473223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a66326164373830306362336165356163323036336133646236656463353666633765613937356333623833646437636662643161396230643934373665333532222c0a2020202020202273697a65223a203438302c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022616d643634222c0a2020202020202020226f73223a20226c696e7578220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a65383932373161323938353037646436366132663131363832346366393432373335623363616331336366376461303435613636326133623131343665353935222c0a2020202020202273697a65223a203438302c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a202261726d3634222c0a2020202020202020226f73223a20226c696e7578220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a66643739306430353633653739336630663733366362636161366434353033346132343032656438653665643035383537633964663039313364393464346438222c0a2020202020202273697a65223a203833382c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022766e642e646f636b65722e7265666572656e63652e646967657374223a20227368613235363a66326164373830306362336165356163323036336133646236656463353666633765613937356333623833646437636662643161396230643934373665333532222c0a202020202020202022766e642e646f636b65722e7265666572656e63652e74797065223a20226174746573746174696f6e2d6d616e6966657374220a2020202020207d2c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022756e6b6e6f776e222c0a2020202020202020226f73223a2022756e6b6e6f776e220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a31343039373437333132356430373463346435313936326134643366323361373035653662643735653862376365323838363731626132613839616562323930222c0a2020202020202273697a65223a203833382c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022766e642e646f636b65722e7265666572656e63652e646967657374223a20227368613235363a65383932373161323938353037646436366132663131363832346366393432373335623363616331336366376461303435613636326133623131343665353935222c0a202020202020202022766e642e646f636b65722e7265666572656e63652e74797065223a20226174746573746174696f6e2d6d616e6966657374220a2020202020207d2c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022756e6b6e6f776e222c0a2020202020202020226f73223a2022756e6b6e6f776e220a2020202020207d0a202020207d0a20205d0a7d', NULL, NULL, 'ImageIndex', 1710256613690, 0, 0, NULL, 1710256613690, 1710256613690, 0),
  (6, 1, 1, 'sha256:33f521b4a6f9e719af668f1fe19d9e94f00399cfd19f71517b1454717d2a796d', 838, 82526, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a30653462613264393439396136356433373161613739633931613833623362336337363434663232643134623734383236393665666663633036663737383431222c0a202020202273697a65223a203234310a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a39623937363562333831643138316336303335373031386563333835383139333465626335353635353837333032356330333039313166613139353563366364222c0a2020202020202273697a65223a2038313030312c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f737064782e6465762f446f63756d656e74220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a30353366396263626332633939383063623364383965643431323833663333636431326233636530303336303738363937636135363239376239396335356262222c0a2020202020202273697a65223a20313238342c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f736c73612e6465762f70726f76656e616e63652f76302e32220a2020202020207d0a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22756e6b6e6f776e222c226f73223a22756e6b6e6f776e222c22636f6e666967223a7b7d2c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a39623937363562333831643138316336303335373031386563333835383139333465626335353635353837333032356330333039313166613139353563366364222c227368613235363a30353366396263626332633939383063623364383965643431323833663333636431326233636530303336303738363937636135363239376239396335356262225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Provenance', 1710256627084, 0, 0, NULL, 1710256627084, 1710256627084, 0),
  (7, 1, 1, 'sha256:9322cd57e6cdbcf101d37438d97df528c45285ebb7455c9bef264404ccc09a31', 480, 3259790, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a37393364316562636633623261393464613734653365623963643834323662353530643762643061396331353036616661343137623234346537626266646662222c0a202020202273697a65223a203630300a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6c617965722e76312e7461722b677a6970222c0a20202020202022646967657374223a20227368613235363a32363164613431363236373362393365356330653737303061333731386434306263633038366462663234623165633962353462636130623832333030363236222c0a2020202020202273697a65223a20333235393139300a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a2261726d3634222c22636f6e666967223a7b22456e76223a5b22504154483d2f7573722f6c6f63616c2f7362696e3a2f7573722f6c6f63616c2f62696e3a2f7573722f7362696e3a2f7573722f62696e3a2f7362696e3a2f62696e225d2c22436d64223a5b222f62696e2f7368225d2c224f6e4275696c64223a6e756c6c7d2c2263726561746564223a22323032322d31312d32325432323a33393a32312e3137363439303930355a222c22686973746f7279223a5b7b2263726561746564223a22323032322d31312d32325432323a33393a32312e3033333937303431335a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f7029204144442066696c653a3638356235656461646631643562663061656232616563333566383130643833383736653664326561303930336232313366373561396335663064633539303120696e202f20227d2c7b2263726561746564223a22323032322d31312d32325432323a33393a32312e3137363439303930355a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f70292020434d44205b5c222f62696e2f73685c225d222c22656d7074795f6c61796572223a747275657d5d2c226f73223a226c696e7578222c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a31623537376138666238636532353032336130656330613137613664633364366161396363613938396637353435373830306362353531373965653265383334225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Image', 1710256627098, 0, 0, NULL, 1710256627098, 1710256627098, 0),
  (8, 1, 1, 'sha256:2d277fcdab280dfbeefb044ee9ab8b5d2609308b305812012d797cf943f2626b', 480, 3371306, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a34303266316165313365613537386464633034643739373933313762636233313533633462326465383337346363346162343930643339346662316565376664222c0a202020202273697a65223a203630300a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6c617965722e76312e7461722b677a6970222c0a20202020202022646967657374223a20227368613235363a63313538393837623035353137623666326335393133663361636566316632313832613332333435613330346665333537653361636535666164636164373135222c0a2020202020202273697a65223a20333337303730360a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22616d643634222c22636f6e666967223a7b22456e76223a5b22504154483d2f7573722f6c6f63616c2f7362696e3a2f7573722f6c6f63616c2f62696e3a2f7573722f7362696e3a2f7573722f62696e3a2f7362696e3a2f62696e225d2c22436d64223a5b222f62696e2f7368225d2c224f6e4275696c64223a6e756c6c7d2c2263726561746564223a22323032322d31312d32325432323a31393a32392e3030383536323332365a222c22686973746f7279223a5b7b2263726561746564223a22323032322d31312d32325432323a31393a32382e3837303830313835355a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f7029204144442066696c653a3538376361653731393639383731643363363435366438343461383739356466396236346231326337313063323735323935613131383262343666363330653720696e202f20227d2c7b2263726561746564223a22323032322d31312d32325432323a31393a32392e3030383536323332365a222c22637265617465645f6279223a222f62696e2f7368202d632023286e6f70292020434d44205b5c222f62696e2f73685c225d222c22656d7074795f6c61796572223a747275657d5d2c226f73223a226c696e7578222c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a64656437613232306262303538653238656533323534666262613034636139306236373930373034323434323437363161353361303433623933623631326266225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Image', 1710256627124, 0, 0, NULL, 1710256627124, 1710256627124, 0),
  (9, 1, 1, 'sha256:7b1f92419ec4a75c303343acc0393a6becc9c374f519bf38cd9ed4e0e676227b', 838, 82544, 'application/vnd.oci.image.manifest.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a202022636f6e666967223a207b0a20202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e636f6e6669672e76312b6a736f6e222c0a2020202022646967657374223a20227368613235363a35633266323463353537383934383063653238613836306533633539656439346463643161373463323335663830623535653930633464636330376564373863222c0a202020202273697a65223a203234310a20207d2c0a2020226c6179657273223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a66613366303839666239346439333731613165383363643961323834393933363134626437343330303063636164393865323461363534396365643433353530222c0a2020202020202273697a65223a2038313031392c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f737064782e6465762f446f63756d656e74220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e696e2d746f746f2b6a736f6e222c0a20202020202022646967657374223a20227368613235363a65393436633136393734396236616437616133396465353139306632663432393734633230353634386337356264326666346438343732633330613264353363222c0a2020202020202273697a65223a20313238342c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022696e2d746f746f2e696f2f7072656469636174652d74797065223a202268747470733a2f2f736c73612e6465762f70726f76656e616e63652f76302e32220a2020202020207d0a202020207d0a20205d0a7d', '\x7b22617263686974656374757265223a22756e6b6e6f776e222c226f73223a22756e6b6e6f776e222c22636f6e666967223a7b7d2c22726f6f746673223a7b2274797065223a226c6179657273222c22646966665f696473223a5b227368613235363a66613366303839666239346439333731613165383363643961323834393933363134626437343330303063636164393865323461363534396365643433353530222c227368613235363a65393436633136393734396236616437616133396465353139306632663432393734633230353634386337356264326666346438343732633330613264353363225d7d7d', 'application/vnd.oci.image.config.v1+json', 'Provenance', 1710256627144, 0, 0, NULL, 1710256627144, 1710256627144, 0),
  (10, 1, 1, 'sha256:34a192d1baa2929229d64f3ec33650efa608eadcdd3314f6984c846f71e39ace', 1607, 2636, 'application/vnd.oci.image.index.v1+json', '\x7b0a202022736368656d6156657273696f6e223a20322c0a2020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e696e6465782e76312b6a736f6e222c0a2020226d616e696665737473223a205b0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a32643237376663646162323830646662656566623034346565396162386235643236303933303862333035383132303132643739376366393433663236323662222c0a2020202020202273697a65223a203438302c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022616d643634222c0a2020202020202020226f73223a20226c696e7578220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a39333232636435376536636462636631303164333734333864393764663532386334353238356562623734353563396265663236343430346363633039613331222c0a2020202020202273697a65223a203438302c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a202261726d3634222c0a2020202020202020226f73223a20226c696e7578220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a33336635323162346136663965373139616636363866316665313964396539346630303339396366643139663731353137623134353437313764326137393664222c0a2020202020202273697a65223a203833382c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022766e642e646f636b65722e7265666572656e63652e646967657374223a20227368613235363a32643237376663646162323830646662656566623034346565396162386235643236303933303862333035383132303132643739376366393433663236323662222c0a202020202020202022766e642e646f636b65722e7265666572656e63652e74797065223a20226174746573746174696f6e2d6d616e6966657374220a2020202020207d2c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022756e6b6e6f776e222c0a2020202020202020226f73223a2022756e6b6e6f776e220a2020202020207d0a202020207d2c0a202020207b0a202020202020226d6564696154797065223a20226170706c69636174696f6e2f766e642e6f63692e696d6167652e6d616e69666573742e76312b6a736f6e222c0a20202020202022646967657374223a20227368613235363a37623166393234313965633461373563333033333433616363303339336136626563633963333734663531396266333863643965643465306536373632323762222c0a2020202020202273697a65223a203833382c0a20202020202022616e6e6f746174696f6e73223a207b0a202020202020202022766e642e646f636b65722e7265666572656e63652e646967657374223a20227368613235363a39333232636435376536636462636631303164333734333864393764663532386334353238356562623734353563396265663236343430346363633039613331222c0a202020202020202022766e642e646f636b65722e7265666572656e63652e74797065223a20226174746573746174696f6e2d6d616e6966657374220a2020202020207d2c0a20202020202022706c6174666f726d223a207b0a202020202020202022617263686974656374757265223a2022756e6b6e6f776e222c0a2020202020202020226f73223a2022756e6b6e6f776e220a2020202020207d0a202020207d0a20205d0a7d', NULL, NULL, 'ImageIndex', 1710256627160, 0, 0, NULL, 1710256627160, 1710256627160, 0);

INSERT INTO "artifact_artifacts" ("artifact_id", "artifact_sub_id")
  VALUES (5, 1),
  (5, 2),
  (5, 3),
  (5, 4),
  (10, 6),
  (10, 7),
  (10, 8),
  (10, 9);

INSERT INTO "tags" ("id", "repository_id", "artifact_id", "name", "pushed_at", "last_pull", "pull_times", "created_at", "updated_at", "deleted_at")
  VALUES (1, 1, 10, '3.18.0', 1710256613698, 1710256627150, 1, 1710256613698, 1710256627168, 0);

