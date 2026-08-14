## gologgen

### install

    git clone https://github.com/rpcox/gologgen.git
    cd gologgen
    go build

### tested on

- Macbook Pro (Tahoe 26.6.1) [arm64]
- RHEL 9.8 [amd]
- Ubuntu 24.04.4 LTS [amd]

### usage

Send 12345 BSD formatted syslog records per worker (3) downrange. The record will appear as

    <134>Aug  9 18:34:42 forge.lan loggen[7356]: JGUVCCWMWSUYATNXIQTNMKAKDXJCIAKUVMUFHSEUFGQFHGIELMBRBJUCCLTTLTBHDEOLVDCBTMWRJXCLWOFTJLVBJEUUPOCAQXMLUQEIXSNQRMTLJANPQMMTMSRTMBIT


    > loggen bsd -dst 192.168.1.122  -count 12345 -w 3 -dport 6000 -stats
     Starting: 2026-08-10 01:34:42.844600
      Elapsed: 43.315458ms
    
    Worker         Emit   Duration             Bytes   Eps
    tcp-worker-01  12345  00d 00:00:00.041557  2148030 297057
    tcp-worker-02  12345  00d 00:00:00.041609  2148030 296686
    tcp-worker-03  12345  00d 00:00:00.042110  2148030 293161
    
               AVG 12345  00d 00:00:00.041759  2148030 295624
             TOTAL 37035                       6444090


Send 12345 IETF formatted syslog records per worker (3) downrange. The record will appear as

    <134>1 2026-08-09T18:30:18.043Z forge.lan loggen 7338 - - CACDXBPAGWWHUUCSDTXCWEGDJWJUYYEBELDGGSPDGRHVRGUCFKMWYMPWIFGIAGOBWIIALFCWMWNVOOHSRJUIBKFILYKLPDIOOLYFGABEEVOMBSXCHQEQIFHLJMYKPYYI


    > loggen ietf -dst 192.168.1.122  -count 12345 -w 3 -dport 6000 -stats
     Starting: 2026-08-10 01:30:17.989031
      Elapsed: 99.228125ms
    
    Worker         Emit   Duration             Bytes   Eps
    tcp-worker-01  12345  00d 00:00:00.097770  2308515 126265
    tcp-worker-02  12345  00d 00:00:00.052601  2308515 234690
    tcp-worker-03  12345  00d 00:00:00.040510  2308515 304733
    
               AVG 12345  00d 00:00:00.063627  2308515 194020
             TOTAL 37035                       6925545


### notes

    -tls option currently not implemented



