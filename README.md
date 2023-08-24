## gologgen

Generates RFC3164 and RFC5424 syslog records.  The messages are user selecte length random strings.


### Generate RFC 3164 records

Default priority is 'local0.info' and the default message length is 64 byte

    gologgen -rfc3164 -tcp -server loghost

As received at the destination

    <134>Aug 24 00:30:15 spud gologgen: VKEKFNVRWVOHJREBFKUEPLQRYLHVODSOLCRJVABCBGWUQQHWAMADURJEMWLAJYKK

### Generate RFC 3164 records with RFC 3339 timestamp

    gologgen -rfc3164 -rfc3339 -tcp -server loghost

As received at the destination

    <134>2023-08-24T00:35:40Z spud gologgen: RVJXCPFBCGUKYPLRRLNECNHLLUNNCODTWKIFBJEPYJLNWLVJYDKCUXJKJEFXFBYK

### Generate default RFC 5424 records

    gologgen -tcp -server 192.168.0.253

As received at the destination

    <134>1 2023-08-24T12:38:15.92Z Rics-MBP gologgen - - - VHQBLXNPCIJMSOPHKSKQCOCDAAIGERQTYVVKWQLPRDIQSNHNOLFBRQFAOBDSKKTT

### Generate RFC 5424 records with MSGID

    gologgen -tcp -server 192.168.0.253 -msgid MYID

As received at the destination

    <134>1 2023-08-24T12:38:06.99Z Rics-MBP gologgen - MYID - EUOPTALBYVVFWXAAQKMSLCOFURERYDWTLGMTUVPAEIDTBAPJVQMYNGCATVXHKUAM

### Generate RFC 5424 records with MSGID and PROCID (PID)

    gologgen -tcp -server 192.168.0.253 -msgid MYID -procid

As received at the destination

    <134>1 2023-08-24T12:39:15.17Z Rics-MBP gologgen 93300 MYID - YDQGDMADSJAYAQFJBNNILTKOBIBDBXXJESVVHRODNQASQHIGMEUCGKNFFKLLNXPM

### Generate RFC 5424 records with MSGID, PROCID and structured data

    gologgen -tcp -server 192.168.0.253 -msgid MYID -procid -sd "[exampleSDID@32473 iut=\"3\" eventSource=\"Application\" eventID=\"1011\"]"

As received at the destination

    <134>1 2023-08-24T12:42:33.08Z Rics-MBP gologgen 93340 MYID [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] QFCWTLDNWRJMDLFQYNFXHIGEUFLWOODAOJSISGSHDFWMSXCTVVPJSPALJXAKJXSI
