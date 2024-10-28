#!/bin/bash
touch compare.txt

verdict_path="./verdict.txt"
output_path="./output.txt"
answer_path="./answer.txt"
compare_path="./compare.txt"

touch output.txt
truncate -s 0 $verdict_path $compare_path
go build main.go
stat=$?
if [ $stat -eq 1 ]; then
    echo "Compilation Error" >> $verdict_path
    echo "compilation error bruh"
    exit 1
fi
timeout 2s ./main < input.txt > output.txt
exit_status=$?

if [ $exit_status -eq 124 ]; then
    echo "Time Limit Exceeded" >> $verdict_path

elif [ $exit_status -ne 0 ]; then
    echo "Runtime Error" >> $verdict_path

else
    # removed the output.txt ? heh, not so fast
    if [ ! -f $output_path ]; then
        touch $output_path
    fi

    diff --brief $output_path $answer_path > $compare_path

    if [ -s $compare_path ]; then
        echo "Wrong Answer" >> $verdict_path
    else
        echo "Accepted" >> $verdict_path
    fi
fi

cat $verdict_path
