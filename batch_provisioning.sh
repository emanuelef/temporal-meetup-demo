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

# Generate a random MAC address
device_mac=$(generate_mac)

# Print the generated MAC address
echo "Generated MAC address: $device_mac"

# JSON payload with the randomized MAC address
payload="{ \"name\": \"guestNetwork\", \"deviceMac\": \"$device_mac\" }"

# Send the POST request using curl
curl -X POST http://localhost:8080/provision \
    -H "Content-Type: application/json" \
    -d "$payload"

# Loop to run curl 100 times with a sleep of 100 ms in between
for ((i = 1; i <= 100; i++)); do
    # Generate a random MAC address
    device_mac=$(generate_mac)

    # Print the generated MAC address
    echo "Generated MAC address: $device_mac"

    # JSON payload with the randomized MAC address
    payload="{ \"name\": \"guestNetwork\", \"deviceMac\": \"$device_mac\" }"

    # Send the POST request using curl
    curl -X POST http://localhost:8080/provision \
        -H "Content-Type: application/json" \
        -d "$payload"

    # Sleep for 100 milliseconds
    sleep 0.1
done

device_mac="FF:BB:CC:11:11:77"

# Print the specified MAC address
echo "Specified MAC address: $device_mac"

# JSON payload with the specified MAC address
payload="{ \"name\": \"guestNetwork\", \"deviceMac\": \"$device_mac\" }"

# Send the POST request using curl
curl -X POST http://localhost:8080/provision \
    -H "Content-Type: application/json" \
    -d "$payload"

# Loop to run curl 100 times with a sleep of 100 ms in between
for ((i = 1; i <= 100; i++)); do
    # Generate a random MAC address
    device_mac=$(generate_mac)

    # Print the generated MAC address
    echo "Generated MAC address: $device_mac"

    # JSON payload with the randomized MAC address
    payload="{ \"name\": \"guestNetwork\", \"deviceMac\": \"$device_mac\" }"

    # Send the POST request using curl
    curl -X POST http://localhost:8080/provision \
        -H "Content-Type: application/json" \
        -d "$payload"

    # Sleep for 100 milliseconds
    sleep 0.1
done
