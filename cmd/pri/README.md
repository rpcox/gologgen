## pri


    NAME
	pri
	
    SYNOPSIS
	pri INTEGER | STRING
	pri list
	
    DESCRIPTION
	See RFC 5424. pri will calculate the integer PRI value from a dotted string
	value in the format facility.severity, or from the dotted string value, pri 
	will return the integer value.

      list
	Show the list of valid facilities and severities
	
    EXAMPLES
	> pri 134
	local0.info
	>
	
	pri audit.info
	110
	>


