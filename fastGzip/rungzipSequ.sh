find /tmp/test -name *.gz -delete
fastgzip -DOP=1 /tmp/test ".*tar$"
