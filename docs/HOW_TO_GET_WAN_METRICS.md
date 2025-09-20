# steps to get full stat for wan metrics on fiber connection.


## Steps
- curl 'http://192.168.2.254/ws/NeMo/Intf/lan:getMIBs' \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US,en;q=0.9' \
  -H 'Connection: keep-alive' \
  -b '121adbc0/accept-language=en-US,en; 121adbc0/sessid=d1beL30i1eeEKaEnbLzUAC1j' \
  -H 'Origin: http://192.168.2.254' \
  -H 'Referer: http://192.168.2.254/' \
  -H 'Sec-GPC: 1' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36' \
  -H 'authorization: X-Sah cXYm6EpCjmJlFF3uvpjLD4o7+itD1UI0W+7AddzGcVvayRy8l/zJVno2k8QaxhdX' \
  -H 'content-type: application/x-sah-ws-4-call+json' \
  -H 'x-context: cXYm6EpCjmJlFF3uvpjLD4o7+itD1UI0W+7AddzGcVvayRy8l/zJVno2k8QaxhdX' \
  --data-raw '{"service":"NeMo.Intf.eth0","method":"get","parameters":{}}' \
  --insecure

the respose is:
{"status":{"Name":"eth0","Enable":true,"Status":true,"Flags":"enabled netdev eth bcmeth physical netdev-monitor statmon ipv4 ipv6 wan netdev-bound netdev-up up ipv6-up","Alias":"cpe-eth0","NATEnabled":false,"NetDevIndex":9,"NetDevType":"ether","NetDevFlags":"up broadcast allmulti multicast","NetDevName":"eth0","LLAddress":"88:D2:74:AB:05:D1","TxQueueLen":1000,"MTU":1508,"NetDevState":"up","IPv4Forwarding":true,"IPv4ForceIGMPVersion":0,"IPv4AcceptSourceRoute":true,"IPv4AcceptLocal":false,"IPv4AcceptRedirects":false,"IPv4ArpFilter":false,"IPv6AcceptRA":true,"IPv6ActAsRouter":false,"IPv6AutoConf":true,"IPv6MaxRtrSolicitations":3,"IPv6RtrSolicitationInterval":4000,"IPv6AcceptSourceRoute":false,"IPv6AcceptRedirects":true,"IPv6OptimisticDAD":false,"IPv6AcceptDAD":2,"IPv6Disable":false,"IPv6HostPart":"","RtTable":0,"RtPriority":0,"IPv6AddrDelegate":"","LastChangeTime":74,"LastChange":413736,"CurrentBitRate":1000,"MaxBitRateSupported":1000,"MaxBitRateEnabled":-1,"CurrentDuplexMode":"Full","DuplexModeEnabled":"Auto","PowerSavingSupported":true,"PowerSavingEnabled":false,"PowerSavingStatus":"Inactive","IPv6RouterDownTimeout":0,"PhysicalInterface":"Ethernet"}}

- curl 'http://192.168.2.254/ws/NeMo/Intf/lan:getMIBs' \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US,en;q=0.9' \
  -H 'Connection: keep-alive' \
  -b '121adbc0/accept-language=en-US,en; 121adbc0/sessid=d1beL30i1eeEKaEnbLzUAC1j' \
  -H 'Origin: http://192.168.2.254' \
  -H 'Referer: http://192.168.2.254/' \
  -H 'Sec-GPC: 1' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36' \
  -H 'authorization: X-Sah cXYm6EpCjmJlFF3uvpjLD4o7+itD1UI0W+7AddzGcVvayRy8l/zJVno2k8QaxhdX' \
  -H 'content-type: application/x-sah-ws-4-call+json' \
  -H 'x-context: cXYm6EpCjmJlFF3uvpjLD4o7+itD1UI0W+7AddzGcVvayRy8l/zJVno2k8QaxhdX' \
  --data-raw '{"service":"NeMo.Intf.eth0","method":"getNetDevStats","parameters":{}}' \
  --insecure

the respose is: 
{"status":{"RxPackets":140045735,"TxPackets":53436844,"RxBytes":7460867824,"TxBytes":8258926237,"RxErrors":0,"TxErrors":0,"RxDropped":0,"TxDropped":0,"Multicast":2238353,"Collisions":0,"RxLengthErrors":0,"RxOverErrors":0,"RxCrcErrors":0,"RxFrameErrors":0,"RxFifoErrors":0,"RxMissedErrors":0,"TxAbortedErrors":0,"TxCarrierErrors":0,"TxFifoErrors":0,"TxHeartbeatErrors":0,"TxWindowErrors":0}}