
for month in {01..06}; do
    for day in {01..31}; do
        for hour in {1..23}; do
            if [ -r "2025-$month-$day-$hour.json.gz" ]; then
                echo "already exists $month-$day-$hour"
            else
                echo "Downlaoding $month-$day-$hour"
                timeout $1s wget "https://data.gharchive.org/2025-$month-$day-$hour.json.gz" >> log.txt
            fi
        done
    done
done
