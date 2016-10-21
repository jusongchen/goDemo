find /tmp/test -name *.gz -delete
if [-z %1] ; then
fastgzip -TOS=%1 /tmp/test ".*tar$";
else
fastgzip  /tmp/test ".*tar$";
fi 
