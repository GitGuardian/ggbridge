#!/bin/bash

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
mkdir -p "logs_$TIMESTAMP"

for i in 0 1 2; do
  for c in nginx ggbridge; do
    echo -e "=== INDEX: $i - CONTAINER: $c - $(date) ===\n" > "logs_$TIMESTAMP/index-${i}_${c}.log"
    kubectl logs -l "app.kubernetes.io/instance=ggbridge,index=$i" -c "$c" -n ggbridge --tail=1000 2>&1 | sed 's/\x1b\[[0-9;]*[mGKHF]//g' >> "logs_$TIMESTAMP/index-${i}_${c}.log"
  done
done

tar -czf "ggbridge_logs_$TIMESTAMP.tgz" "logs_$TIMESTAMP"
rm -rf "logs_$TIMESTAMP"
echo "✅ Archive created: ggbridge_logs_$TIMESTAMP.tgz"