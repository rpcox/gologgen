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

    gologgen -tcp -server loghost

As received at the destination

    <134>1 2023-08-24T12:38:15.92Z spud gologgen - - - VHQBLXNPCIJMSOPHKSKQCOCDAAIGERQTYVVKWQLPRDIQSNHNOLFBRQFAOBDSKKTT

### Generate RFC 5424 records with MSGID

    gologgen -tcp -server loghost -msgid MYID

As received at the destination

    <134>1 2023-08-24T12:38:06.99Z spud gologgen - MYID - EUOPTALBYVVFWXAAQKMSLCOFURERYDWTLGMTUVPAEIDTBAPJVQMYNGCATVXHKUAM

### Generate RFC 5424 records with MSGID and PROCID (PID)

    gologgen -tcp -server loghost -msgid MYID -procid

As received at the destination

    <134>1 2023-08-24T12:39:15.17Z spud gologgen 93300 MYID - YDQGDMADSJAYAQFJBNNILTKOBIBDBXXJESVVHRODNQASQHIGMEUCGKNFFKLLNXPM

### Generate RFC 5424 records with MSGID, PROCID and structured data

    gologgen -tcp -server loghost -msgid MYID -procid -sd "[exampleSDID@32473 iut=\"3\" eventSource=\"Application\" eventID=\"1011\"]"

As received at the destination

    <134>1 2023-08-24T12:42:33.08Z spud gologgen 93340 MYID [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] QFCWTLDNWRJMDLFQYNFXHIGEUFLWOODAOJSISGSHDFWMSXCTVVPJSPALJXAKJXSI
