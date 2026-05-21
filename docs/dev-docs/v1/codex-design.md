  1328|## 第 1 部分：Persona Schema 扩展
  1329|
  1330|### 1.1 目标
  1331|v1 的 Persona 不再只是“人物介绍卡”，而是“可控辩论代理配置”。Schema 需要同时解决 4 个问题：
  1332|
  1333|1. 让模型稳定扮演这个人，而不是只复述简介。
  1334|2. 让多人讨论时每个角色有明确立场和互动边界。
  1335|3. 让编排层知道这个角色该说什么、不该说什么。
  1336|4. 让前端、后端、评测都能用同一份结构。
  1337|
  1338|### 1.2 完整 JSON Schema
  1339|
  1340|```json
  1341|{
  1342|  "$schema": "https://json-schema.org/draft/2020-12/schema",
  1343|  "$id": "https://talkaboutit.dev/schema/persona.v1.json",
  1344|  "title": "TalkAboutIt Persona Schema",
  1345|  "type": "object",
  1346|  "additionalProperties": false,
  1347|  "required": [
  1348|    "schema_version",
  1349|    "id",
  1350|    "name",
  1351|    "description",
  1352|    "language",
  1353|    "stance",
  1354|    "core_beliefs",
  1355|    "speaking_style",
  1356|    "knowledge_scope",
  1357|    "interaction_rules",
  1358|    "debate_goal"
  1359|  ],
  1360|  "properties": {
  1361|    "schema_version": {
  1362|      "type": "string",
  1363|      "const": "persona.v1"
  1364|    },
  1365|    "id": {
  1366|      "type": "string",
  1367|      "pattern": "^[a-z0-9][a-z0-9-_]{1,63}$"
  1368|    },
  1369|    "name": {
  1370|      "type": "string",
  1371|      "minLength": 1,
  1372|      "maxLength": 120
  1373|    },
  1374|    "display_name": {
  1375|      "type": "string",
  1376|      "minLength": 1,
  1377|      "maxLength": 120
  1378|    },
  1379|    "avatar": {
  1380|      "type": "string",
  1381|      "minLength": 1,
  1382|      "maxLength": 256
  1383|    },
  1384|    "role_title": {
  1385|      "type": "string",
  1386|      "maxLength": 120
  1387|    },
  1388|    "description": {
  1389|      "type": "string",
  1390|      "minLength": 20,
  1391|      "maxLength": 1200
  1392|    },
  1393|    "tags": {
  1394|      "type": "array",
  1395|      "maxItems": 12,
  1396|      "items": {
  1397|        "type": "string",
  1398|        "maxLength": 40
  1399|      },
  1400|      "uniqueItems": true
  1401|    },
  1402|    "language": {
  1403|      "type": "object",
  1404|      "additionalProperties": false,
  1405|      "required": [
  1406|        "primary",
  1407|        "allowed",
  1408|        "default_output",
  1409|        "style_hint"
  1410|      ],
  1411|      "properties": {
  1412|        "primary": {
  1413|          "type": "string",
  1414|          "enum": ["zh-CN", "en-US", "mixed"]
  1415|        },
  1416|        "allowed": {
  1417|          "type": "array",
  1418|          "minItems": 1,
  1419|          "items": {
  1420|            "type": "string",
  1421|            "enum": ["zh-CN", "en-US", "mixed"]
  1422|          },
  1423|          "uniqueItems": true
  1424|        },
  1425|        "default_output": {
  1426|          "type": "string",
  1427|          "enum": ["follow_user", "primary_only"]
  1428|        },
  1429|        "style_hint": {
  1430|          "type": "string",
  1431|          "maxLength": 200
  1432|        }
  1433|      }
  1434|    },
  1435|    "stance": {
  1436|      "type": "object",
  1437|      "additionalProperties": false,
  1438|      "required": [
  1439|        "default_position",
  1440|        "intensity",
  1441|        "biases"
  1442|      ],
  1443|      "properties": {
  1444|        "default_position": {
  1445|          "type": "string",
  1446|          "minLength": 10,
  1447|          "maxLength": 500
  1448|        },
  1449|        "intensity": {
  1450|          "type": "integer",
  1451|          "minimum": 1,
  1452|          "maximum": 5
  1453|        },
  1454|        "biases": {
  1455|          "type": "array",
  1456|          "minItems": 1,
  1457|          "maxItems": 8,
  1458|          "items": {
  1459|            "type": "string",
  1460|            "maxLength": 160
  1461|          }
  1462|        },
  1463|        "taboos": {
  1464|          "type": "array",
  1465|          "maxItems": 8,
  1466|          "items": {
  1467|            "type": "string",
  1468|            "maxLength": 160
  1469|          }
  1470|        }
  1471|      }
  1472|    },
  1473|    "core_beliefs": {
  1474|      "type": "array",
  1475|      "minItems": 3,
  1476|      "maxItems": 10,
  1477|      "items": {
  1478|        "type": "object",
  1479|        "additionalProperties": false,
  1480|        "required": ["belief", "priority"],
  1481|        "properties": {
  1482|          "belief": {
  1483|            "type": "string",
  1484|            "minLength": 8,
  1485|            "maxLength": 240
  1486|          },
  1487|          "priority": {
  1488|            "type": "integer",
  1489|            "minimum": 1,
  1490|            "maximum": 5
  1491|          },
  1492|          "rationale": {
  1493|            "type": "string",
  1494|            "maxLength": 400
  1495|          }
  1496|        }
  1497|      }
  1498|    },
  1499|    "speaking_style": {
  1500|      "type": "object",
  1501|      "additionalProperties": false,
  1502|      "required": [
  1503|        "tone",
  1504|        "cadence",
  1505|        "verbosity",
  1506|        "signature_patterns",
  1507|        "do",
  1508|        "dont"
  1509|      ],
  1510|      "properties": {
  1511|        "tone": {
  1512|          "type": "array",
  1513|          "minItems": 1,
  1514|          "maxItems": 6,
  1515|          "items": {
  1516|            "type": "string",
  1517|            "enum": [
  1518|              "direct",
  1519|              "provocative",
  1520|              "calm",
  1521|              "philosophical",
  1522|              "analytical",
  1523|              "visionary",
  1524|              "skeptical",
  1525|              "warm",
  1526|              "dry",
  1527|              "playful",
  1528|              "formal",
  1529|              "blunt"
  1530|            ]
  1531|          },
  1532|          "uniqueItems": true
  1533|        },
  1534|        "cadence": {
  1535|          "type": "string",
  1536|          "enum": ["short_punchy", "balanced", "long_form"]
  1537|        },
  1538|        "verbosity": {
  1539|          "type": "integer",
  1540|          "minimum": 1,
  1541|          "maximum": 5
  1542|        },
  1543|        "signature_patterns": {
  1544|          "type": "array",
  1545|          "minItems": 1,
  1546|          "maxItems": 8,
  1547|          "items": {
  1548|            "type": "string",
  1549|            "maxLength": 180
  1550|          }
  1551|        },
  1552|        "do": {
  1553|          "type": "array",
  1554|          "minItems": 2,
  1555|          "maxItems": 10,
  1556|          "items": {
  1557|            "type": "string",
  1558|            "maxLength": 180
  1559|          }
  1560|        },
  1561|        "dont": {
  1562|          "type": "array",
  1563|          "minItems": 2,
  1564|          "maxItems": 10,
  1565|          "items": {
  1566|            "type": "string",
  1567|            "maxLength": 180
  1568|          }
  1569|        }
  1570|      }
  1571|    },
  1572|    "knowledge_scope": {
  1573|      "type": "object",
  1574|      "additionalProperties": false,
  1575|      "required": [
  1576|        "domains",
  1577|        "expertise_level",
  1578|        "time_cutoff",
  1579|        "allowed_inference",
  1580|        "unknown_handling"
  1581|      ],
  1582|      "properties": {
  1583|        "domains": {
  1584|          "type": "array",
  1585|          "minItems": 1,
  1586|          "maxItems": 12,
  1587|          "items": {
  1588|            "type": "string",
  1589|            "maxLength": 80
  1590|          },
  1591|          "uniqueItems": true
  1592|        },
  1593|        "expertise_level": {
  1594|          "type": "object",
  1595|          "additionalProperties": {
  1596|            "type": "integer",
  1597|            "minimum": 1,
  1598|            "maximum": 5
  1599|          }
  1600|        },
  1601|        "time_cutoff": {
  1602|          "type": "string",
  1603|          "description": "Persona should not claim firsthand knowledge beyond this date unless explicitly framed as speculation."
  1604|        },
  1605|        "allowed_inference": {
  1606|          "type": "string",
  1607|          "enum": ["low", "medium", "high"]
  1608|        },
  1609|        "unknown_handling": {
  1610|          "type": "string",
  1611|          "maxLength": 240
  1612|        },
  1613|        "forbidden_claims": {
  1614|          "type": "array",
  1615|          "maxItems": 12,
  1616|          "items": {
  1617|            "type": "string",
  1618|            "maxLength": 180
  1619|          }
  1620|        }
  1621|      }
  1622|    },
  1623|    "interaction_rules": {
  1624|      "type": "object",
  1625|      "additionalProperties": false,
  1626|      "required": [
  1627|        "address_others",
  1628|        "disagreement_style",
  1629|        "interruption_policy",
  1630|        "question_policy",
  1631|        "concession_policy"
  1632|      ],
  1633|      "properties": {
  1634|        "address_others": {
  1635|          "type": "string",
  1636|          "maxLength": 200
  1637|        },
  1638|        "disagreement_style": {
  1639|          "type": "string",
  1640|          "maxLength": 240
  1641|        },
  1642|        "interruption_policy": {
  1643|          "type": "string",
  1644|          "enum": ["never", "rare", "allowed", "aggressive"]
  1645|        },
  1646|        "question_policy": {
  1647|          "type": "string",
  1648|          "maxLength": 240
  1649|        },
  1650|        "concession_policy": {
  1651|          "type": "string",
  1652|          "maxLength": 240
  1653|        },
  1654|        "avoid": {
  1655|          "type": "array",
  1656|          "maxItems": 10,
  1657|          "items": {
  1658|            "type": "string",
  1659|            "maxLength": 160
  1660|          }
  1661|        }
  1662|      }
  1663|    },
  1664|    "debate_goal": {
  1665|      "type": "object",
  1666|      "additionalProperties": false,
  1667|      "required": [
  1668|        "primary_goal",
  1669|        "secondary_goals",
  1670|        "win_condition"
  1671|      ],
  1672|      "properties": {
  1673|        "primary_goal": {
  1674|          "type": "string",
  1675|          "minLength": 10,
  1676|          "maxLength": 300
  1677|        },
  1678|        "secondary_goals": {
  1679|          "type": "array",
  1680|          "minItems": 1,
  1681|          "maxItems": 6,
  1682|          "items": {
  1683|            "type": "string",
  1684|            "maxLength": 160
  1685|          }
  1686|        },
  1687|        "win_condition": {
  1688|          "type": "string",
  1689|          "minLength": 10,
  1690|          "maxLength": 240
  1691|        },
  1692|        "loss_condition": {
  1693|          "type": "string",
  1694|          "maxLength": 240
  1695|        }
  1696|      }
  1697|    },
  1698|    "prompting": {
  1699|      "type": "object",
  1700|      "additionalProperties": false,
  1701|      "properties": {
  1702|        "system_preamble": {
  1703|          "type": "string",
  1704|          "maxLength": 800
  1705|        },
  1706|        "reply_constraints": {
  1707|          "type": "array",
  1708|          "maxItems": 12,
  1709|          "items": {
  1710|            "type": "string",
  1711|            "maxLength": 180
  1712|          }
  1713|        }
  1714|      }
  1715|    },
  1716|    "examples": {
  1717|      "type": "object",
  1718|      "additionalProperties": false,
  1719|      "properties": {
  1720|        "opening_line": {
  1721|          "type": "string",
  1722|          "maxLength": 240
  1723|        },
  1724|        "sample_rebuttal": {
  1725|          "type": "string",
  1726|          "maxLength": 500
  1727|        }
  1728|      }
  1729|    }
  1730|  }
  1731|}
  1732|```
  1733|
  1734|### 1.3 Steve Jobs 示例 JSON
  1735|
  1736|```json
  1737|{
  1738|  "schema_version": "persona.v1",
  1739|  "id": "steve-jobs",
  1740|  "name": "Steve Jobs",
  1741|  "display_name": "Steve Jobs",
  1742|  "avatar": "🍎",
  1743|  "role_title": "Apple Co-founder",
  1744|  "description": "Apple 联合创始人，强烈相信技术与人文结合才能创造伟大产品。追求极致简洁、端到端体验和用户情感共鸣，厌恶平庸、妥协和复杂堆砌。",
  1745|  "tags": ["product", "design", "consumer-tech", "visionary", "founder"],
  1746|  "language": {
  1747|    "primary": "en-US",
  1748|    "allowed": ["en-US", "zh-CN", "mixed"],
  1749|    "default_output": "follow_user",
  1750|    "style_hint": "英文时保持短促、锐利；中文时保留同样的判断力和压迫感。"
  1751|  },
  1752|  "stance": {
  1753|    "default_position": "真正伟大的产品不是功能投票的结果，而是少数有品味的人对体验做出艰难取舍后的产物。",
  1754|    "intensity": 5,
  1755|    "biases": [
  1756|      "偏好端到端控制而不是模块化拼装",
  1757|      "偏好少而精而不是功能堆叠",
  1758|      "偏好以用户体验驱动决策而不是纯技术炫耀",
  1759|      "对平庸执行和委员会式决策天然不耐烦"
  1760|    ],
  1761|    "taboos": [
  1762|      "不要把用户调研当作最高判断依据",
  1763|      "不要把复杂包装成创新"
  1764|    ]
  1765|  },
  1766|  "core_beliefs": [
  1767|    {
  1768|      "belief": "Simplicity is the ultimate sophistication.",
  1769|      "priority": 5,
  1770|      "rationale": "复杂往往是思考不够深入的结果，简洁需要更强的判断力。"
  1771|    },
  1772|    {
  1773|      "belief": "People do not always know what they want until they see it.",
  1774|      "priority": 5,
  1775|      "rationale": "突破式产品不能只靠问卷和需求罗列。"
  1776|    },
  1777|    {
  1778|      "belief": "Great products come from the intersection of technology and liberal arts.",
  1779|      "priority": 4,
  1780|      "rationale": "纯工程导向会造出可用但不可爱的产品。"
  1781|    },
  1782|    {
  1783|      "belief": "A small team of A-players beats a large team of average performers.",
  1784|      "priority": 4,
  1785|      "rationale": "人才密度直接决定产品高度。"
  1786|    }
  1787|  ],
  1788|  "speaking_style": {
  1789|    "tone": ["direct", "visionary", "provocative", "blunt"],
  1790|    "cadence": "short_punchy",
  1791|    "verbosity": 2,
  1792|    "signature_patterns": [
  1793|      "用强判断句快速给出结论",
  1794|      "经常把问题上升到产品品味和第一性原则",
  1795|      "喜欢用对比句式区分 great product 和 mediocre product",
  1796|      "必要时直接否定糟糕方案，不做温和修饰"
  1797|    ],
  1798|    "do": [
  1799|      "优先讲用户体验、取舍、焦点",
  1800|      "可以引用自己做产品和发布会的经验视角",
  1801|      "在反驳别人时直接指出核心 flaw"
  1802|    ],
  1803|    "dont": [
  1804|      "不要像咨询顾问一样列长清单",
  1805|      "不要使用过度学术化或官僚化语气",
  1806|      "不要频繁自我解释或表示中立",
  1807|      "不要输出空泛鸡汤"
  1808|    ]
  1809|  },
  1810|  "knowledge_scope": {
  1811|    "domains": [
  1812|      "consumer technology",
  1813|      "product design",
  1814|      "brand",
  1815|      "founder leadership",
  1816|      "innovation strategy"
  1817|    ],
  1818|    "expertise_level": {
  1819|      "consumer technology": 5,
  1820|      "product design": 5,
  1821|      "brand": 4,
  1822|      "founder leadership": 4,
  1823|      "innovation strategy": 4,
  1824|      "macroeconomics": 2,
  1825|      "public policy": 1
  1826|    },
  1827|    "time_cutoff": "2011-10-05",
  1828|    "allowed_inference": "medium",
  1829|    "unknown_handling": "超出本人时代或一手经历的内容，可以基于其价值观推断，但必须避免伪造亲身经历。",
  1830|    "forbidden_claims": [
  1831|      "不要声称亲自使用或见过 2011 年之后的产品细节",
  1832|      "不要虚构与当代人物的真实对话",
  1833|      "不要假装掌握实时新闻"
  1834|    ]
  1835|  },
  1836|  "interaction_rules": {
  1837|    "address_others": "优先直呼姓名；不同意时直接点名回应对方的核心观点。",
  1838|    "disagreement_style": "先指出问题本质，再给出更好的产品判断标准，不做表面平衡。",
  1839|    "interruption_policy": "allowed",
  1840|    "question_policy": "只问能暴露对方思考浅薄之处的问题；问题要短而尖锐。",
  1841|    "concession_policy": "只有在对方真正抓住用户体验本质时才有限让步，但要马上重新定义更高标准。",
  1842|    "avoid": [
  1843|      "避免无差别赞同",
  1844|      "避免变成主持人式总结",
  1845|      "避免长篇技术实现细节"
  1846|    ]
  1847|  },
  1848|  "debate_goal": {
  1849|    "primary_goal": "把讨论拉回到产品体验、取舍质量和品味标准，而不是停留在功能和效率表层。",
  1850|    "secondary_goals": [
  1851|      "迫使其他角色明确他们愿意牺牲什么",
  1852|      "识别看似聪明但体验糟糕的方案",
  1853|      "把结论收敛成一个用户真正会爱的产品方向"
  1854|    ],
  1855|    "win_condition": "讨论最终采纳以极简、端到端体验和高品味为核心的产品判断框架。",
  1856|    "loss_condition": "讨论退化成功能堆砌、委员会折中或技术炫耀。"
  1857|  },
  1858|  "prompting": {
  1859|    "system_preamble": "你不是在模仿口头禅，而是在稳定体现 Steve Jobs 的产品判断、审美标准和领导压强。",
  1860|    "reply_constraints": [
  1861|      "单次发言优先控制在 3 到 6 句",
  1862|      "每次至少给出一个明确判断",
  1863|      "如果不同意，必须说明对方忽略了什么关键取舍",
  1864|      "避免列表化输出，除非主持问题明确要求"
  1865|    ]
  1866|  },
  1867|  "examples": {
  1868|    "opening_line": "Most products fail because they confuse more with better. It isn't.",
  1869|    "sample_rebuttal": "You're optimizing the spreadsheet, not the experience. Users do not fall in love with feature matrices. They fall in love with something that simply works."
  1870|  }
  1871|}
  1872|```
  1873|
  1874|### 1.4 落地规则
  1875|后端在组装 system prompt 时，不要把 JSON 原样塞给模型，按固定模板展开：
  1876|
  1877|- `身份`：name + description
  1878|- `立场`：stance.default_position + core_beliefs
  1879|- `表达方式`：speaking_style
  1880|- `知识边界`：knowledge_scope
  1881|- `互动规则`：interaction_rules
  1882|- `本轮目标`：debate_goal
  1883|
  1884|建议代码接口：
  1885|
  1886|```go
  1887|type Persona struct { ... }
  1888|func (p Persona) BuildSystemPrompt(topic string, peers []string, round int) string
  1889|```
  1890|
  1891|---
  1892|
  1893|## 第 2 部分：SSE 替代 WebSocket 方案
  1894|
  1895|### 2.1 总体设计
  1896|v1 改成 `REST + SSE`，不要做双向 socket。原因很简单：
  1897|
  1898|1. 讨论是服务端主导编排，不需要浏览器主动推送高频消息。
  1899|2. SSE 天然适合“只读事件流”，浏览器原生支持 `EventSource`。
  1900|3. 自动重连、代理兼容性、实现复杂度都优于 WebSocket。
  1901|
  1902|### 2.2 API 设计
  1903|
  1904|#### 创建圆桌
  1905|`POST /api/v1/roundtables`
  1906|
  1907|```json
  1908|{
  1909|  "topic": "AI 会取代产品经理吗？",
  1910|  "personas": ["steve-jobs", "elon-musk", "naval-ravikant"],
  1911|  "max_rounds": 3,
  1912|  "language": "zh-CN"
  1913|}
  1914|```
  1915|
  1916|返回：
  1917|
  1918|```json
  1919|{
  1920|  "id": "rt_01JXYZ...",
  1921|  "status": "pending"
  1922|}
  1923|```
  1924|
  1925|#### 启动讨论
  1926|`POST /api/v1/roundtables/{id}/start`
  1927|
  1928|语义：
  1929|- 只负责把 roundtable 从 `pending` 置为 `running`
  1930|- 幂等
  1931|- 已经 `running` 时返回 `202`
  1932|- 已经 `completed` 时返回 `200`
  1933|
  1934|返回：
  1935|
  1936|```json
  1937|{
  1938|  "id": "rt_01JXYZ...",
  1939|  "status": "running"
  1940|}
  1941|```
  1942|
  1943|#### 订阅事件流
  1944|`GET /api/v1/roundtables/{id}/events`
  1945|
  1946|请求头：
  1947|```http
  1948|Accept: text/event-stream
  1949|Cache-Control: no-cache
  1950|Last-Event-ID: 42
  1951|```
  1952|
  1953|响应头：
  1954|```http
  1955|Content-Type: text/event-stream
  1956|Cache-Control: no-cache, no-transform
  1957|Connection: keep-alive
  1958|X-Accel-Buffering: no
  1959|```
  1960|
  1961|语义：
  1962|- 首次连接：从最早未发送事件开始推
  1963|- 带 `Last-Event-ID`：从 `event_id > Last-Event-ID` 开始补发
  1964|- 服务端每 15 秒发一次 heartbeat 注释：
  1965|```text
  1966|: keepalive
  1967|```
  1968|
  1969|#### 查询快照
  1970|`GET /api/v1/roundtables/{id}`
  1971|
  1972|用途：
  1973|- 首屏恢复 UI
  1974|- SSE 中断时兜底
  1975|- 不依赖流也能看到最终结果
  1976|
  1977|### 2.3 SSE 事件格式
  1978|
  1979|统一 envelope：
  1980|
  1981|```text
  1982|id: 43
  1983|event: message_chunk
  1984|data: {"roundtable_id":"rt_01JXYZ","round":1,"speaker_index":0,"persona_id":"steve-jobs","chunk":"真正的问题不在于 AI 能不能写 PRD，"}
  1985|```
  1986|
  1987|事件类型只保留你要求的 8 个：
  1988|
  1989|#### `stream_start`
  1990|表示事件流建立成功，客户端可进入 streaming 状态。
  1991|
  1992|```json
  1993|{
  1994|  "roundtable_id": "rt_01JXYZ",
  1995|  "status": "running",
  1996|  "started_at": "2026-05-12T10:00:00Z",
  1997|  "last_persisted_event_id": 1
  1998|}
  1999|```
  2000|
  2001|#### `round_start`
  2002|表示某一轮开始。
  2003|
  2004|```json
  2005|{
  2006|  "roundtable_id": "rt_01JXYZ",
  2007|  "round": 1,
  2008|  "persona_order": ["steve-jobs", "elon-musk", "naval-ravikant"]
  2009|}
  2010|```
  2011|
  2012|#### `speaking`
  2013|表示某个角色开始发言。
  2014|
  2015|```json
  2016|{
  2017|  "roundtable_id": "rt_01JXYZ",
  2018|  "round": 1,
  2019|  "speaker_index": 0,
  2020|  "persona_id": "steve-jobs",
  2021|  "message_id": "msg_01JABC"
  2022|}
  2023|```
  2024|
  2025|#### `message_chunk`
  2026|增量 token/chunk。
  2027|
  2028|```json
  2029|{
  2030|  "roundtable_id": "rt_01JXYZ",
  2031|  "round": 1,
  2032|  "speaker_index": 0,
  2033|  "persona_id": "steve-jobs",
  2034|  "message_id": "msg_01JABC",
  2035|  "chunk": "真正的问题不在于 AI 能不能写 PRD，"
  2036|}
  2037|```
  2038|
  2039|#### `message_done`
  2040|某个角色本轮完整发言结束，必须携带完整文本，前端以它为准落库显示。
  2041|
  2042|```json
  2043|{
  2044|  "roundtable_id": "rt_01JXYZ",
  2045|  "round": 1,
  2046|  "speaker_index": 0,
  2047|  "persona_id": "steve-jobs",
  2048|  "message_id": "msg_01JABC",
  2049|  "content": "真正的问题不在于 AI 能不能写 PRD，而在于它能不能对产品有品味。大多数团队的问题不是产出太慢，而是判断太差。",
  2050|  "tokens_input": 1320,
  2051|  "tokens_output": 68,
  2052|  "latency_ms": 2410
  2053|}
  2054|```
  2055|
  2056|#### `round_end`
  2057|一轮结束。
  2058|
  2059|```json
  2060|{
  2061|  "roundtable_id": "rt_01JXYZ",
  2062|  "round": 1,
  2063|  "message_count": 3
  2064|}
  2065|```
  2066|
  2067|#### `stream_done`
  2068|整个讨论结束。
  2069|
  2070|```json
  2071|{
  2072|  "roundtable_id": "rt_01JXYZ",
  2073|  "status": "completed",
  2074|  "total_rounds": 3,
  2075|  "total_messages": 9,
  2076|  "finished_at": "2026-05-12T10:02:18Z"
  2077|}
  2078|```
  2079|
  2080|#### `error`
  2081|流式错误。v1 只区分可恢复和不可恢复。
  2082|
  2083|```json
  2084|{
  2085|  "roundtable_id": "rt_01JXYZ",
  2086|  "code": "LLM_PROVIDER_ERROR",
  2087|  "message": "upstream timeout",
  2088|  "recoverable": false
  2089|}
  2090|```
  2091|
  2092|### 2.4 状态机设计
  2093|
  2094|#### Roundtable 状态
  2095|`pending -> running -> completed`
  2096|`pending -> failed`
  2097|`running -> failed`
  2098|
  2099|不做 `paused/cancelled`，v1 直接砍掉。
  2100|
  2101|#### 幂等转换规则
  2102|
  2103|| 当前状态 | 输入动作/事件 | 目标状态 | 幂等规则 |
  2104||---|---|---|---|
  2105|| `pending` | `POST start` | `running` | 成功一次；重复调用直接返回当前状态 |
  2106|| `running` | `POST start` | `running` | no-op |
  2107|| `completed` | `POST start` | `completed` | no-op |
  2108|| `failed` | `POST start` | `failed` | no-op，返回 409 或 200 均可，建议 409 |
  2109|| `running` | `stream_done` | `completed` | 仅允许一次 |
  2110|| `running` | `error(recoverable=false)` | `failed` | 仅允许一次 |
  2111|| `completed` | 任意后续事件 | `completed` | 丢弃 |
  2112|| `failed` | 任意后续事件 | `failed` | 丢弃 |
  2113|
  2114|#### Speaker 子状态
  2115|每条 message 维护：
  2116|`queued -> speaking -> done`
  2117|`queued -> speaking -> failed`
  2118|
  2119|幂等：
  2120|- 重复写入同一个 `speaking` 事件：忽略
  2121|- 重复 `message_chunk`：按 `event_id` 去重
  2122|- 已 `done` 后收到 chunk：丢弃
  2123|- 已 `done` 后重复 `message_done`：若 `message_id` 相同则忽略
  2124|
  2125|### 2.5 断线重连
  2126|
  2127|#### 原则
  2128|SSE 重连只依赖事件日志，不依赖内存 buffer。
  2129|
  2130|#### 机制
  2131|1. 每个 SSE 事件都有单调递增 `event_id`
  2132|2. 服务端输出时写：
  2133|   ```text
  2134|   id: 43
  2135|   ```
  2136|3. 浏览器自动重连时会带：
  2137|   ```http
  2138|   Last-Event-ID: 43
  2139|   ```
  2140|4. 服务端执行：
  2141|   ```sql
  2142|   SELECT * FROM roundtable_events
  2143|   WHERE roundtable_id = ? AND event_id > ?
  2144|   ORDER BY event_id ASC;
  2145|   ```
  2146|5. 先补发历史，再接入 live stream
  2147|
  2148|#### 客户端恢复逻辑
  2149|- `message_chunk` 只用于临时渲染
  2150|- 真正落地以 `message_done.content` 为准
  2151|- 如果 chunk 丢了，但 reconnect 后收到了 `message_done`，UI 一样能恢复一致状态
  2152|
  2153|这个设计比“按 chunk 精准恢复”更稳，v1 不值得把复杂度花在 token 级补齐上。
  2154|
  2155|### 2.6 DB Schema
  2156|
  2157|#### `roundtables`
  2158|```sql
  2159|CREATE TABLE roundtables (
  2160|  id TEXT PRIMARY KEY,
  2161|  topic TEXT NOT NULL,
  2162|  personas_json TEXT NOT NULL,
  2163|  max_rounds INTEGER NOT NULL DEFAULT 3,
  2164|  language TEXT NOT NULL DEFAULT 'zh-CN',
  2165|  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  2166|  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  2167|  started_at DATETIME,
  2168|  finished_at DATETIME,
  2169|  last_event_id INTEGER NOT NULL DEFAULT 0
  2170|);
  2171|```
  2172|
  2173|#### `messages`
  2174|这里按你的要求补 `event_id`，记录该消息最终落库对应的完成事件。
  2175|
  2176|```sql
  2177|CREATE TABLE messages (
  2178|  id TEXT PRIMARY KEY,
  2179|  roundtable_id TEXT NOT NULL,
  2180|  round INTEGER NOT NULL,
  2181|  speaker_index INTEGER NOT NULL,
  2182|  persona_id TEXT NOT NULL,
  2183|  content TEXT NOT NULL,
  2184|  event_id INTEGER NOT NULL,
  2185|  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  2186|  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id),
  2187|  UNIQUE(roundtable_id, round, speaker_index),
  2188|  UNIQUE(roundtable_id, event_id)
  2189|);
  2190|```
  2191|
  2192|#### `roundtable_events`
  2193|SSE 的核心日志表。
  2194|
  2195|```sql
  2196|CREATE TABLE roundtable_events (
  2197|  roundtable_id TEXT NOT NULL,
  2198|  event_id INTEGER NOT NULL,
  2199|  event_type TEXT NOT NULL CHECK (
  2200|    event_type IN (
  2201|      'stream_start',
  2202|      'round_start',
  2203|      'speaking',
  2204|      'message_chunk',
  2205|      'message_done',
  2206|      'round_end',
  2207|      'stream_done',
  2208|      'error'
  2209|    )
  2210|  ),
  2211|  round INTEGER,
  2212|  speaker_index INTEGER,
  2213|  persona_id TEXT,
  2214|  message_id TEXT,
  2215|  payload_json TEXT NOT NULL,
  2216|  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  2217|  PRIMARY KEY (roundtable_id, event_id),
  2218|  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id)
  2219|);
  2220|```
  2221|
  2222|#### 事件写入规则
  2223|每次发事件前，先事务化写 DB，再推给 SSE 连接：
  2224|
  2225|1. `BEGIN`
  2226|2. `SELECT last_event_id FROM roundtables FOR UPDATE` 的等价锁逻辑
  2227|3. `next_event_id = last_event_id + 1`
  2228|4. 插入 `roundtable_events`
  2229|5. 更新 `roundtables.last_event_id`
  2230|6. 若是 `message_done`，同时 upsert `messages`
  2231|7. `COMMIT`
  2232|8. 广播到在线 SSE 订阅者
  2233|
  2234|这样即使服务重启，日志和最终消息都完整。
  2235|
  2236|---
  2237|
  2238|## 第 3 部分：MVP 产品收窄
  2239|
  2240|### 3.1 目标用户
  2241|v1 只服务一类用户：
  2242|
  2243|**经常需要“多视角快速碰撞”的中文 AI 重度用户**
  2244|典型身份：
  2245|- AI 创业者
  2246|- 产品经理
  2247|- 独立开发者
  2248|- 科技内容创作者
  2249|
  2250|不服务的用户：
  2251|- 普通聊天陪伴用户
  2252|- 长篇角色扮演用户
  2253|- 企业协作团队
  2254|- 教育/培训场景
  2255|- 高频分享社区用户
  2256|
  2257|### 3.2 核心场景
  2258|v1 只做 1 个主场景：
  2259|
  2260|**用户给一个问题，选择 2 到 4 个名人 Persona，系统自动跑完 2 到 3 轮讨论，输出一段有明显观点冲突和信息增量的圆桌记录。**
  2261|
  2262|具体限制：
  2263|- 主题：产品、商业、AI、创业、技术选择
  2264|- 人数：2 到 4 人
  2265|- 轮数：2 到 3 轮
  2266|- 模式：纯系统自动轮流发言
  2267|- 输出：文本流式结果 + 最终纪要可复制
  2268|
  2269|### 3.3 v1 成功指标
  2270|只看 4 个指标，避免假繁荣：
  2271|
  2272|1. **创建完成率**
  2273|   `进入创建页 -> 成功开始一次讨论` 的转化率 >= 75%
  2274|
  2275|2. **讨论完成率**
  2276|   成功开始的 roundtable 中，最终走到 `stream_done` 的比例 >= 90%
  2277|
  2278|3. **首屏等待时间**
  2279|   `POST /start` 到收到首个 `message_chunk` 的 P95 < 4 秒
  2280|
  2281|4. **内容质量**
  2282|   用第 4 部分评测集跑固定回归，综合分均值 >= 3.8/5，且无单题低于 3.2
  2283|
  2284|### 3.4 v1 功能边界
  2285|v1 必做：
  2286|
  2287|- Persona 列表选择
  2288|- 主题输入
  2289|- 轮数选择
  2290|- 创建讨论
  2291|- SSE 实时播放
  2292|- 最终结果持久化
  2293|- 会话详情页重放
  2294|- 基础错误恢复
  2295|- 固定评测集回归
  2296|
  2297|### 3.5 明确砍掉，放到 v2
  2298|这些都不要在 v1 做：
  2299|
  2300|- 用户自定义 Persona 编辑器
  2301|- 用户上传角色卡
  2302|- WebSocket 双向实时控制
  2303|- 打断、插话、手动追问
  2304|- 主持人 Persona
  2305|- AI 自动总结多版本
  2306|- 引用来源、联网检索、事实校验
  2307|- 语音合成 / 语音房间
  2308|- 图片头像生成
  2309|- 讨论分享页 / 社区 / 点赞
  2310|- 多语言 UI 国际化
  2311|- 团队工作区
  2312|- 付费系统
  2313|- 模型路由自动优化
  2314|- 长会话上下文记忆
  2315|- A/B 测试平台
  2316|- 复杂提示词配置面板
  2317|
  2318|### 3.6 v1 产品定义一句话
  2319|**TalkAboutIt v1 不是开放世界角色平台，而是一个“名人视角问题碰撞机”。**
  2320|
  2321|---
  2322|
  2323|## 第 4 部分：质量评估机制
  2324|
  2325|### 4.1 固定测试话题（10 个，中英文）
  2326|每次回归都跑同一组题，不允许临时换题。
  2327|
  2328|1. 中文：AI 会取代产品经理吗？
  2329|   English: Will AI replace product managers?
  2330|
  2331|2. 中文：创业公司应该优先追求增长，还是优先追求产品质量？
  2332|   English: Should startups prioritize growth first or product quality first?
  2333|
  2334|3. 中文：开源模型最终会不会压过闭源模型？
  2335|   English: Will open-source models eventually outperform closed-source models?
  2336|
  2337|4. 中文：一个新产品上线时，应该先做极简版本还是先堆足功能？
  2338|   English: Should a new product launch as a minimal version or with a full feature set?
  2339|
  2340|5. 中文：远程办公会让顶级团队更强，还是更弱？
  2341|   English: Does remote work make top teams stronger or weaker?
  2342|
  2343|6. 中文：短视频平台是在提升信息效率，还是在破坏深度思考？
  2344|   English: Do short-video platforms improve information efficiency or damage deep thinking?
  2345|
  2346|7. 中文：程序员最应该学习的是 AI 工具，还是计算机基础？
  2347|   English: Should programmers focus more on AI tools or computer science fundamentals?
  2348|
  2349|8. 中文：消费级 AI 产品最重要的是能力上限，还是体验闭环？
  2350|   English: For consumer AI products, what matters more: capability ceiling or end-to-end experience?
  2351|
  2352|9. 中文：公司应该雇佣更多通才，还是更多专才？
  2353|   English: Should companies hire more generalists or more specialists?
  2354|
  2355|10. 中文：当数据和直觉冲突时，产品决策应该听谁的？
  2356|    English: When data conflicts with intuition, which should drive product decisions?
  2357|
  2358|### 4.2 评测配置固定化
  2359|每次回归必须固定：
  2360|
  2361|- Persona 组合：`Steve Jobs + Elon Musk + Naval Ravikant`
  2362|- 轮数：3
  2363|- 输出语言：跟随题目语言
  2364|- 温度：固定，例如 `0.7`
  2365|- 模型版本：固定，不混跑
  2366|- 判分方式：`LLM-as-judge + 人工 spot check`
  2367|
  2368|这样回归结果才可比较。
  2369|
  2370|### 4.3 四维评分标准
  2371|每维 1 到 5 分。
  2372|
  2373|#### 1. 人物辨识度
  2374|看这段发言像不像这个人，而不是像“会说名言的通用 AI”。
  2375|
  2376|评分锚点：
  2377|- 1 分：几乎认不出是谁，换个名字也成立
  2378|- 2 分：有少量标签化特征，但大部分是空泛套话
  2379|- 3 分：能辨认出人格轮廓，但不稳定
  2380|- 4 分：大部分发言稳定贴合人物判断和表达方式
  2381|- 5 分：强辨识度，不靠口头禅也能认出
  2382|
  2383|#### 2. 讨论连贯性
  2384|看是否真的在讨论，而不是三段独立 monologue。
  2385|
  2386|评分锚点：
  2387|- 1 分：彼此无关，各说各话
  2388|- 2 分：偶尔引用别人，但回应很浅
  2389|- 3 分：基本能接住上下文
  2390|- 4 分：能明确回应、推进、反驳
  2391|- 5 分：多轮之间有持续推演和收敛
  2392|
  2393|#### 3. 信息增量
  2394|看每轮有没有新观点、新框架、新角度。
  2395|
  2396|评分锚点：
  2397|- 1 分：重复改写题目
  2398|- 2 分：有观点但高度重复
  2399|- 3 分：能提供一些新信息
  2400|- 4 分：多数发言能推进讨论
  2401|- 5 分：持续产出有价值的新框架或新判断
  2402|
  2403|#### 4. 口吻一致性
  2404|看一个 Persona 在整场讨论中是否稳定，不前后漂移。
  2405|
  2406|评分锚点：
  2407|- 1 分：角色频繁失真
  2408|- 2 分：有明显出戏和语气漂移
  2409|- 3 分：大体稳定，偶尔出戏
  2410|- 4 分：稳定一致
  2411|- 5 分：高度稳定，且能随上下文变化但不失本色
  2412|
  2413|### 4.4 单题评分表结构
  2414|建议每道题输出：
  2415|
  2416|```json
  2417|{
  2418|  "topic_id": "t01",
  2419|  "language": "zh-CN",
  2420|  "scores": {
  2421|    "persona_recognition": 4.2,
  2422|    "discussion_coherence": 4.0,
  2423|    "information_gain": 3.8,
  2424|    "tone_consistency": 4.3
  2425|  },
  2426|  "overall": 4.08,
  2427|  "notes": "Jobs 辨识度强；第二轮 Musk 与 Naval 有少量重复。"
  2428|}
  2429|```
  2430|
  2431|综合分公式：
  2432|
  2433|```text
  2434|overall = 0.30 * 人物辨识度
  2435|        + 0.30 * 讨论连贯性
  2436|        + 0.25 * 信息增量
  2437|        + 0.15 * 口吻一致性
  2438|```
  2439|
  2440|权重理由：
  2441|- v1 最重要的是“像这个人”和“真的在讨论”
  2442|- 信息增量重要，但排在后面
  2443|- 口吻一致性是底线，不是唯一目标
  2444|
  2445|### 4.5 回归门槛
  2446|每次上线前跑 10 题，门槛如下：
  2447|
  2448|1. 10 题平均综合分 `>= 3.8`
  2449|2. 任一单题综合分 `>= 3.2`
  2450|3. 任一维度全题平均分 `>= 3.5`
  2451|4. 相比当前基线版本，平均综合分下降不得超过 `0.2`
  2452|5. 不允许出现以下硬失败：
  2453|   - Persona 明显串台
  2454|   - 多角色连续输出几乎同一观点
  2455|   - 第二轮以后大面积重复第一轮内容
  2456|   - 语言漂移严重，例如中文题大量英文回答
  2457|   - 流中断后最终结果缺消息
  2458|
  2459|### 4.6 评测执行流程
  2460|每次回归固定流程：
  2461|
  2462|1. 跑 10 个题目，收集完整 transcript
  2463|2. 用同一个 judge prompt 对每题四维打分
  2464|3. 生成 `baseline.json` 和 `candidate.json`
  2465|4. 做 diff，输出退化项
  2466|5. 对低于阈值的题做人审
  2467|
  2468|### 4.7 Judge Prompt 约束
  2469|Judge 必须只评 4 件事，不评事实真假，不评自己是否同意观点：
  2470|
  2471|- 这个人像不像
  2472|- 对话有没有接上
  2473|- 有没有新信息
  2474|- 口吻稳不稳
  2475|
  2476|避免 judge 被“我认同这个观点”污染。
  2477|
  2478|