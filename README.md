# caddy-nats (EXPERIMENTAL)

> Run `nats-server` as a [caddy app](https://caddyserver.com/docs/extending-caddy#app-modules) with experimental oauth2 authentication.

## Example usage

- First build the project:

```bash
go build ./cmd/caddy
```

- Allow caddy to bind to port 80 and 443:

```bash
sudo setcap cap_net_bind_service=+ep ./caddy
```

- Update the example config with the oauth2 endpoint of your choice. Configuration should be a valid [Endpoint struct](https://github.com/charbonnierg/caddy-nats/blob/rewrite/oauthproxy/endpoint.go#L21).

  ⚠ example config won't run without change, because it uses Azure Provider with fake data.

Start the example:

```bash
./caddy run -c example.json
```

- Visit `https://localhost`. You should be redirected to configured OAuth2 provider to authenticate. Once authentication is succesfull, you should be redirected back to `https://localhost` and see metrics displayed in the page.

- Open developer tools and checkout cookies. Copy the value of `_oauth2_proxy_0` and `_oauth2_proxy_1` cookies.

- Open a terminal, connect using APP account:

```bash
nats pub foo bar --user APP --password "_oauth2_proxy_0=<value>;_oauth2_proxy_1=<value>"
```

Connect using SYS account:

```bash
nats server ls --user SYS --password "_oauth2_proxy_0=<value>;_oauth2_proxy_1=<value>"
```

> Note: At the moment, user is granted access for ANY account as long as OAuth2 session state can be decoded successfully
> from cookies.

<details>

- Client side:

```bash
> .\nats.exe pub foo bar --server tls://localhost --user APP --password "_oauth2_proxy_0=XhcPNJBIOLa4lc3AiGaD0yYwePPmbk6rfEBFcCt320RZRWDrU330_-SLkW1iDp8NIklG1IM_pBk1exBXi5nXTbHSmtYXF89LAtlAT4yzb0NVIT8nIZbcZ7J00O-FCYZlady37P_L9D2hzercH-TQ7SWBaFAR9zkGoao_7-uHn-2SWqOFaa7PIqLQvZ8UCmtnl0XEX9e1kHkWdQ23n12rDUiyik-p7E8jqaEG2E8ahNTDMpnawnWGLO6BB93sMLjDtGWWjlMkPJfoTfEtJZWkA7IZr14prClXMfru6BHy4rbBkWQUnTiS9Eu249ubizij0aU99S8wnEdEdrlbyoWfHwzkZZEaZf0SiApZdMZMbZqeGybJ409b38o4S9BWco0k2pYs_PLLbtweyguKG_HwMpcGin-mmEtV0vcXsbRbBmMcI218DkgHiv9O3arHt9P2I2E3ooYBGs5gQGwngoSHthNmci3WOsKWMFpjwTXD2BdeVtUlLxXKodhJK4BO1VHXbW_JYsrO0sDwbv0zPnDyScFuvIVUGXzKvMF_ycPVXB0KPaIG69nyJ7RBhDwTayK9wf8Zk-6S1rcAGScr73sBUDc0v_-DJh9bDIfwusX5_CzMcJVOBCnizLasv8BrNAqH_B5jmTL35D-UD9EzpWbseVywORACTlC8Xy-m1rhcTswZm150fel0eE3LVVLbPHNsjz1EO4iWj_lrbGioMdueKlw1Zi5b9FGiaeJVhqUIn8T_E6B257Cdot1y8lMGh6sCsdTHFtts5D4v5G81Sqxd0lyxIHyJ3Tt2PfPbrb0DVh35rS14nHezO7h8gkQIiZ9dzxMffqfO9youwFOmYZmLHrX2h_sdhSRW-mDZL_2fdNJicoE3s2oVzq4SDf-cjBuX_EKYQaiKEIar0nHVwCsEcZRoBxx1zyHLdXxZbRuKsPMtvYUXO0vzUJ6ooZx_Qm2AHk5zWxkWq6ihUMk2_BKUpa8aa8SmXpWmupI_QCEzHzv79g5fa3KB9qI2d4ODCV3mAaSyeNyCuIyGoKKI_0ucupYwAf1GUdfXRc1U4BgH-jVeGV1TOHkbp-UJWR65HymRP6Eyv6fOkAuR1OLZskpOslce61WGbw04hZTqnHIdAyPH57MIN1yMWtlKxs83Vnw-0yatKwGIewjXFxHkWCWUqW_7mXmSoXF9LYa3XTaBODZVFSo4HephYgHljAQ1kJN0JT5HF0moKtEvXkuoPcypgXtHqKhbPkQu3kbBVUnZOupN9ZwHkGodyIiBDGqvItT-X6H95CfAV3RQABbpsfF9hgUlgjVy8qk2bvMLGkoTd8cNiepMzHrC-dpE_3N3BykDAmNg_lvXcpAjRwPOMqKqjuCoZ5gs27_oX88hLK2UKm6F1QQ7dETUQ79exddDPCrhprXUrGz2awXo_oDNJszuHzaZ7Xd-Yh0Ok5tH4IW0lUt4y7dxRciLTKM_YUxcazRj8AJiYiXkM96vbAkQFQc66hUUH9M6KlUKxvOyL5u2GGlYxgI6D8332HxC0unHLvL7GrTYUfEuZ0OF17HwUoRHGnHSKx-w-cM8Qe6aUveYpwRiKv6JEHzCJvtPxd0I6rC9cOiIfWtiNQyA4l4gFMT9oBwl32TSZCq5iY_NAzA2lmzVSrvAddcYUI2mgq8BvajfhaqHBLmR8_mJq9Du1NXNPLqA8ry5MwC6gpOrufke3JGOp-ZLiqT5ulTmuN8T-2Bw2j45aDfG2qDO7zLaun1-U3o94OpBJgwPeeItTNpxEG8876f0yPNpgZY21L2kIO3866BuwlVWIfW9kZUQ2CXq-GQjM7HwI904CgpVE2hNR4tCSWHCdVrlgRD0tZhMN5fzR93uN6EIxL2DU1m35quRNEhfxuQJjn-L8LdbZPDu0sqgcFtb8-k9qEsWZRNLvTa0fYD_Qn83a9-gK14hZs9yXzG5HaoTrApJ678H_hfmi40bdXUuPBAvBz1Bud8Fj_q6sEyB2YUOXFty0Nu8OS8Ruuma7fmEu7hEMcwNv_j254JESHlE6sSqB4TwBk9xcbIQ-jWMe1OwIyhXv48Tk2bF4VtGxGkKv3Xjck8gd_wE-X-9t0yFDxxdpGobjb7zpO79vmcK8W8BHx3TXWo8Qap8_E3-R08oi-pOwy_VkZqx_kHOKiuTByJx_BXlwRJeoCa9a9j9IkUeWdshO1l54VtXfrRAMp8WurIiS27Nvg_2fNTy7lcHjSXw2ApI7ArMXf1LjzcmejPoeMTDKzgvm998I09H_SvYM7jvfJqXkqdQTuwmyBcAMSJwHZpIOBxFyWX9rjCe7j8cZTgANM4INDzRtYrZUuh8hgKQ16jMF01gbcuashbnXhp75yHhTvz-le9b-WaZExxsR66KfdsJd_gFBVHXRZDIEsvaVpBEzD8NBKwj9JPnl530ncgVXQJ-JOzQKvgf42iSfmE34mJ5FAyqJpTBsc5vrM5azETbgv2Tsh2tKLWSaZUCnZjnM9sxPBcETsTx4hIPttGEkDXee_NsP-dl_imIwwDjpM_r5IowTigeOjL6Ww6tDX5Ki4BPUs1X8zPWVwYaXys4XVmn8FbrBnv5M5kVkDYTIQFtg_PMirUm3hjhbnsnmxDRG36z0va8wSaSKzx9EyTIbgbjAc7T9_vXl40t_MvxMHwOxA8pLCQDom-0EDZbmc_r6HDAV-ch5yXpAyCvwXzc0QWv9q71Jj9Tcm1t1-oGmUDZv1fWBwfXCT8S9AffbRzLB8XhEF-wI3L0PFrQpPV78JwWFAHzevcZk-cBEhgG5NcgPtNXbPVbzTm2d3WKbvuVaH6RGaEHg75U4FqIK1ngYT2xpYJOZfmqqB8zxczDtqjxZlzSZrVCv3l2JY6fwliR00dfJivZwkYoVJaz46rlzZmLUgGEHk0IwSwj1qIFxYsWgCDygWuoJ86uuPqD7SiT_Aku3k374OE7epWAqBooToEO9s10OSOMJrkhHS3ZDc2OOPyfsdMDH7uAg_hZ8gahDnrte1PvXEgkj5O3qLeRMBDnmFsNFKpqcGBWA4-k__4ZoSL1QB128ztm6aM1TG8SXTnewV1VFUchv27XmFpNkXVbZF1846knbXBijd6saSBJNSt3ZerM1eQQcGT5yGsLhUdggMKI4c6P96bb1z1Zf66JMbOJJQ69uvdKsbwc9tJGAhMqlgpGqL1UBUxlz5-0r0ren6RHXrRWyGCJMWQuKs5pBELQViamYLjBDiUXyV50WYW9phPZnkw-8Y5n2NwRurBJcaqfqmsuCnTH6g3uoctjpmCw_PVGbtP5bZCHpot4IZXa3LIXUvzm_ZVUxKgRvyd3wGDZCthFRT1JPlAH2b12H0utHigxkseTYX4qozm3ocr3XPiGCsmW_07GO-efqg0iFCmt1RBR_hMFnO3rG1ob6JVa8oCXbxGfwxeGAIpMk821sUWccrIDOdRJ-8dFrhjj-cnbJDc69MJAQ_zf3b-z_DAgZK6Z7ZcuJMiFYo2BpiZqutTHorZhzAvpBABc-JLhRZMJm3EqjPvVB6Z2CL7y3GksIoef8Aa2k-r5t_qXPVcd2ZlJulF0pApTvVz2XCtJ0A2jgEDa0q5lelmu2rgsm8ucEE4cuoLwprPHWta7QLkr4nbBeSrIKLoSt2FHWOIDJ5E9F6gAAZVLg34JpfFk3Bed9vuzZGtoq-U_XC7jj4DbQ9A6gE61OsVdiCXehLIgY4Mcc3T1CRCCy9zxVqoMRW33cNTc9meFk5hlL8YEBbiOx8TtIGxrKP4sZ61qopm4g2TyfpEv4fsQR8PHEZ0ugYYdYkPKy8mz_b_A7Mbn_fQLeZS_1Fs9xYxkGZyF9r7dGehNlpyANd5hPkU2YlWYOhgsniARqB13NwyhWqNJX2Mw-t6PnmJjSkg;_oauth2_proxy_1=ZfT58OTXK9qo27zwmtwSq8pFbOEEqU1U_INKDsOvDIJE51HjhhXVNPZVtpCh773z2eVaEKRpvJ5goOjZ8TQT2dSf6hOd7lgsZQI62PlWIU2XMuaOcu_J_g4ZDdC8ux0CWUOEtgKXXuiuGxN-dRY_pbE6vw-EvK_BIWIOKV8ePe_JrpKI2A-ylOl-RUXH3bn8kVVDyukHb5-SNIF4_5a8FIW_IohHRyUVP0j3h3qdhfHSMW9qVQejD|1697311652|FUX-7SVnJxN3Od71Y-TzLGAqUdhezsIp6ieNVnwvsQQ="
21:29:32 Published 3 bytes to "foo"
```

- Server side:

```bash
2023/10/14 19:29:32.448 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - Client connection created
2023/10/14 19:29:32.449 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - Starting TLS client connection handshake
2023/10/14 19:29:32.450 DEBUG   events  event   {"name": "tls_get_certificate", "id": "db51c68e-0313-45d4-890d-13528520a565", "origin": "tls", "data": {"client_hello":{"CipherSuites":[49195,49199,49196,49200,52393,52392,49161,49171,49162,49172,156,157,47,53,49170,10,4865,4866,4867],"ServerName":"localhost","SupportedCurves":[29,23,24,25],"SupportedPoints":"AA==","SignatureSchemes":[2052,1027,2055,2053,2054,1025,1281,1537,1283,1539,513,515],"SupportedProtos":null,"SupportedVersions":[772,771],"Conn":{}}}}
2023/10/14 19:29:32.451 DEBUG   tls.handshake   choosing certificate    {"identifier": "localhost", "num_choices": 1}
2023/10/14 19:29:32.452 DEBUG   tls.handshake   default certificate selection results   {"identifier": "localhost", "subjects": ["localhost"], "managed": true, "issuer_key": "local", "hash": "0f14dc388dc9f7bd4c05fcb7f21f1ddc804d35e65e5695446aa32990ae24a55d"}
2023/10/14 19:29:32.452 DEBUG   tls.handshake   matched certificate in cache    {"remote_ip": "127.0.0.1", "remote_port": "16435", "subjects": ["localhost"], "managed": true, "expiration": "2023/10/15 06:57:12.000", "hash": "0f14dc388dc9f7bd4c05fcb7f21f1ddc804d35e65e5695446aa32990ae24a55d"}
2023/10/14 19:29:32.517 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - TLS handshake complete
2023/10/14 19:29:32.517 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - TLS version 1.3, cipher suite TLS_AES_128_GCM_SHA256
2023/10/14 19:29:32.519 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - <<- [CONNECT {"verbose":false,"pedantic":false,"user":"APP","pass":"[REDACTED]","tls_required":true,"name":"NATS CLI Version 0.1.1","lang":"go","version":"1.30.0","protocol":1,"echo":true,"headers":true,"no_responders":true}]
2023/10/14 19:29:32.519 DEBUG   nats.server     ACCOUNT - <<- [PUB $SYS.REQ.USER.AUTH $SYS._INBOX.ZAM5CE1d.OuZGFxpp 6874]
2023/10/14 19:29:32.520 DEBUG   nats.server     ACCOUNT - <<- MSG_PAYLOAD: [omited]
2023/10/14 19:29:32.521 DEBUG   nats.server     pipe - cid:7 - ->> [MSG $SYS.REQ.USER.AUTH 1 $SYS._INBOX.ZAM5CE1d.OuZGFxpp 6874]
2023/10/14 19:29:32.521 DEBUG   nats.auth_callout       Received authorization request  {"payload": "<omited>"}
2023/10/14 19:29:32.523 INFO    oauth2.az       decoding session state: [omited]
2023/10/14 19:29:32.524 DEBUG   nats.server     pipe - cid:7 - <<- [PUB $SYS._INBOX.ZAM5CE1d.OuZGFxpp 1476]
2023/10/14 19:29:32.524 DEBUG   nats.server     pipe - cid:7 - <<- MSG_PAYLOAD: [omited]   <====== This is the auth response from auth callout service
2023/10/14 19:29:32.525 DEBUG   nats.server     ACCOUNT - <-> [DELSUB 6]
2023/10/14 19:29:32.525 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - <<- [PING]
2023/10/14 19:29:32.525 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - ->> [PONG]
2023/10/14 19:29:32.576 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - <<- [PUB foo 3]
2023/10/14 19:29:32.576 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - <<- MSG_PAYLOAD: ["bar"]
2023/10/14 19:29:32.576 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - <<- [PING]
2023/10/14 19:29:32.576 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - ->> [PONG]
2023/10/14 19:29:32.577 DEBUG   nats.server     127.0.0.1:16435 - cid:8 - Client connection closed: Client Closed
```

</details>

### Caddyfile

No support for caddyfile at the moment.

### JSON file

Checkout the file [example.json](./example.json) to see how to configure an NATS server with TLS certificates managed by caddy and auth callout service running as caddy module.

## Next steps

- Use replacers to avoid writing signing key in config
- Add tests
- Add Caddyfile support
- Add auth callout modules (maybe a module validating ID tokens provided by users in connect options ❔)
