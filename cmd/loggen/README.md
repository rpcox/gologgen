## gologgen

Generates RFC3164 and RFC5424 syslog records.  The messages are user selecte length random strings.


### Generate RFC 3164 records

Send one BSD formatted record over TCP to 192.168.1.122:6000 with process name 'spud' and a 64 byte message 

    > loggen bsd -dst 192.168.1.122 -dport 6000 -tag spud -msglen 64

As received at the destination

    <134>Aug  9 18:48:45 forge.lan spud[7419]: FRMDSFTSODFSNMKTHOGIPFKXGDKNCVGKFMTILKCSUICKAKGSXUXLVQGOHFWMBLKE

### Generate default RFC 5424 records

Send one IETF formatted record over TCP to 192.168.1.122:6000 with process name 'spud' and a 64 byte message 

    > loggen ietf -dst 192.168.1.122 -dport 6000 -appname spud -msgid MSGID -msglen 64

As received at the destination

    <134>1 2026-08-09T18:53:34.415Z forge.lan spud 7424 MSGID - XPYIIAYFMLOWTLDABSDMRCAYFJXGLWVMAJDMQVXQDYGOYBPVDLXRTXGRSCKLAVLN

### Publish a record type for a duration of 1 minute and 30 seconds using 3 workers

    > loggen ietf -dst 192.168.1.122  -duration 1m30s -w 3 -dport 6000

### View Stats

Send an IETF formatted record over TCP to 192.168.1.122:6000 with process name 'spud' and a 64 byte message. Send 123 records per worker and check the stats

    > loggen ietf -dst 192.168.1.122 -dport 6000 -appname spud -msgid MSGID -msglen 64 -w 3 -count 123 -stats
     Starting: 2026-08-10 01:55:58.895705
      Elapsed: 3.475333ms
    
    Worker         Emit   Duration             Bytes Eps
    tcp-worker-01  123    00d 00:00:00.001930  15375 63725
    tcp-worker-02  123    00d 00:00:00.001804  15375 68158
    tcp-worker-03  123    00d 00:00:00.001849  15375 66521
    
               AVG 123    00d 00:00:00.001861  15375 66084
             TOTAL 369                         46125

Note: Eps = events (records) per second generated at the source
