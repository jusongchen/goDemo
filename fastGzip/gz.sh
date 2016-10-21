find /tmp/test -name *.gz -delete
if [ -z "$1" ] ; then
fastgzip /tmp/test ".*tar$";
else
fastgzip -DOP=$1 /tmp/test ".*tar$";
fi 
