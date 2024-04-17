#!/bin/bash

# Function to generate a random MAC address
generate_mac() {
    # Generate six random hex digits
    mac=$(printf "%02x" $((RANDOM % 256)))
    for i in {2..5}; do
        mac+=:$(printf "%02x" $((RANDOM % 256)))
    done
    echo "$mac"
}

# Function to send a POST request with specified MAC address
send_post_request() {
    local mac_address="$1"
    local payload="{ \"name\": \"guestNetwork\", \"deviceMac\": \"$mac_address\" }"
    curl -X POST http://localhost:8080/provision \
        -H "Content-Type: application/json" \
        -d "$payload"
    echo
}

# Loop to run curl 100 times with a sleep of 100 ms in between
for ((i = 1; i <= 100; i++)); do
    device_mac=$(generate_mac)
    echo "Generated MAC address: $device_mac"
    send_post_request "$device_mac"
    sleep 0.3
done

# Specify a MAC address
device_mac="FF:BB:CC:11:11:77"

# Print the specified MAC address
echo "Specified MAC address: $device_mac"

# Send the POST request with the specified MAC address
send_post_request "$device_mac"

# Loop to run curl 150 times with a sleep of 400 ms in between
for ((i = 1; i <= 150; i++)); do
    device_mac=$(generate_mac)
    if ((i % 7 == 0)); then
        device_mac="FF:BB:CC:11:11:77" # Set the specified MAC address
    fi
    echo "Generated MAC address: $device_mac"
    send_post_request "$device_mac"
    sleep 0.4
done
