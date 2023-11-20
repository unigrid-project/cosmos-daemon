import json
import time
from colorama import Fore, init
import pexpect
import threading

init(autoreset=True)

DAEMON_PATH = "/home/evan/go/bin/paxd"
LOG_FILE = "stress_test.log"
FROM_ADDRESS = "unigrid1attpyzmtq9k4r42gkglnx0edgukjh969jrsmn8"
TO_ADDRESS = "unigrid1xeq4qwyhxfukx0xyultta0r882ev86jjs4yvtc"

# Define a global variable for the transaction count
tx_count = 0
tx_count_lock = threading.Lock()


def send_tokens(password):
    cmd = f"{DAEMON_PATH} tx bank send {FROM_ADDRESS} {TO_ADDRESS} 1ugd --home=/home/evan/.unigrid-testnet-1 --fees=0.025uugd"
    child = pexpect.spawn(cmd)

    # Always expect the confirmation prompt and confirm
    child.expect(
        "confirm transaction before signing and broadcasting \[y/N\]:")
    child.sendline("y")

    # Check if the password prompt appears
    i = child.expect(["Password:", pexpect.EOF])
    if i == 0:  # Password prompt appeared
        child.sendline(password)
        child.expect(pexpect.EOF)

    # Parse the output to check for transaction success
    output = child.before.decode()
    # print("Transaction Output:", output)  # Debugging statement
    if "txhash" in output:
        return True

    return False


def stress_test(password, duration=10):
    start_time = time.time()
    tx_count = 0
    speeds = []

    with open(LOG_FILE, "w") as log:
        while time.time() - start_time < duration:
            if send_tokens(password):
                tx_count += 1
            current_speed = tx_count / (time.time() - start_time)
            speeds.append(current_speed)
            avg_speed = sum(speeds) / len(speeds)
            slowdown = avg_speed - current_speed

            print(f"{Fore.GREEN}Speed: {current_speed:.2f} tx/s", end=" ")
            print(f"{Fore.YELLOW}Average: {avg_speed:.2f} tx/s", end=" ")
            print(f"{Fore.RED}Slowdown: {slowdown:.2f} tx/s")

            log.write(
                f"{time.time() - start_time:.2f}s: Speed: {current_speed:.2f} tx/s, Average: {avg_speed:.2f} tx/s, Slowdown: {slowdown:.2f} tx/s\n")

        print(f"\nTest completed. Total transactions: {tx_count}")


def display_metrics(duration):
    start_time = time.time()
    speeds = []
    while time.time() - start_time < duration:
        time.sleep(1)  # Update metrics every second
        current_speed = tx_count / (time.time() - start_time)
        speeds.append(current_speed)
        avg_speed = sum(speeds) / len(speeds)
        slowdown = avg_speed - current_speed

        print(f"{Fore.GREEN}Speed: {current_speed:.2f} tx/s", end=" ")
        print(f"{Fore.YELLOW}Average: {avg_speed:.2f} tx/s", end=" ")
        print(f"{Fore.RED}Slowdown: {slowdown:.2f} tx/s")


def send_tokens_threaded(password, duration):
    global tx_count
    start_time = time.time()
    while time.time() - start_time < duration:
        if send_tokens(password):
            with tx_count_lock:
                tx_count += 1


def stress_test_concurrent(password, duration=10, num_threads=30):
    global tx_count
    tx_count = 0  # Reset the transaction count

    # Start threads for sending transactions
    threads = []
    for _ in range(num_threads):
        t = threading.Thread(target=send_tokens_threaded,
                             args=(password, duration))
        threads.append(t)
        t.start()

    # Start a separate thread for displaying metrics
    metrics_thread = threading.Thread(target=display_metrics, args=(duration,))
    metrics_thread.start()

    for t in threads:
        t.join()

    metrics_thread.join()

    print(f"\nTest completed. Total transactions: {tx_count}")


if __name__ == "__main__":
    password = input("Enter the password: ")
    # stress_test(password)
    stress_test_concurrent(password)
