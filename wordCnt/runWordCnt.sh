
dir="/Users/jusong.chen/work/src"
file=".*[.]go$"
pattern='func '

if [ -z "$1" ] ; then
cmd="wordCnt -e=$pattern $dir $file"
else
cmd="wordCnt -e=$pattern -DOP=$1 $dir $file"
fi 
echo $cmd
eval $cmd
