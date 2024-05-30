# to use this script, run the following command:
# python testing.py <path_to_hedgehog_bin>
# this script requires ignite cli to be installed
# sudo apt install python3-termcolor
# sudo apt install python3-requests
# you can monitor what was added to hedgehog in postman or using curl
# https://127.0.0.1:40005/gridspork/mint-storage
# https://127.0.0.1:40005/gridspork/vesting-storage/

import os
import sys
import subprocess
import time
import json
import requests
import urllib3
from datetime import datetime, timedelta, timezone
import random
from termcolor import colored

# Disable InsecureRequestWarning
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def execute_command(command, log_file=None):
    print(colored(f"Executing: {command}", "blue"))
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
    if log_file:
        with open(log_file, 'a') as f:
            f.write(result.stdout)
            f.write(result.stderr)
    else:
        print(f"Output: {result.stdout.strip()}")
    return result.stdout.strip()

def kill_processes(process_name):
    try:
        result = subprocess.run(["ps", "aux"], capture_output=True, text=True)
        current_pid = os.getpid()
        for line in result.stdout.splitlines():
            if process_name in line and "grep" not in line and str(current_pid) not in line:
                pid = int(line.split()[1])
                print(f"Killing process {pid}")
                subprocess.run(["kill", "-9", str(pid)])
                time.sleep(2)  # Ensure the process has time to terminate
    except Exception as e:
        print(f"No {process_name} process was found to kill. Proceeding...")

def get_json_field(json_string, field):
    data = json.loads(json_string)
    return data.get(field)

def get_current_block_height():
    response = requests.get("http://127.0.0.1:26657/block")
    return response.json()['result']['block']['header']['height']

def wait_for_ignite_chain(timeout=60):
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            response = requests.get("http://127.0.0.1:26657/status")
            if response.status_code == 200:
                print(colored("Ignite chain is up and running.", "green"))
                return True
        except requests.exceptions.ConnectionError:
            pass
        time.sleep(5)
    print(colored("Timeout while waiting for Ignite chain to start.", "red"))
    return False

def verify_mint_storage(rest_port):
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork/mint-storage", verify=False)
    return response.json()

def verify_vesting_storage(rest_port):
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork/vesting-storage", verify=False)
    return response.json()

def verify_account_vesting(account_address):
    response = requests.get(f"http://127.0.0.1:1317/cosmos/auth/v1beta1/accounts/{account_address}")
    return response.json()

if len(sys.argv) < 2:
    print("Usage: python test_suite.py <path_to_hedgehog_bin>")
    sys.exit(1)

hedgehog_bin = sys.argv[1]
rest_port = 40005  # Hardcoded Hedgehog URL

# Step 1: Remove existing Unigrid local data
print(colored("Step 1: Removing existing Unigrid local data...", "yellow"))
execute_command('rm -rf ~/.local/share/unigrid')

# Step 2: Generate a new key pair
print(colored(f"Step 2: Generating a new key pair using {hedgehog_bin}...", "yellow"))
key_gen_output = execute_command(f'{hedgehog_bin} util key-generate')
lines = key_gen_output.splitlines()
private_key = next(line.split(': ')[1] for line in lines if "Private Key" in line)
public_key = next(line.split(': ')[1] for line in lines if "Public Key" in line)

# Ensure only one set of keys is used throughout the script
generated_private_key = private_key
generated_public_key = public_key

# Step 3: Kill any running Hedgehog and Ignite processes
print(colored("Step 3: Killing any running Hedgehog and Ignite processes...", "yellow"))
kill_processes("hedgehog-0.0.8")
kill_processes("ignite")

print(colored("Starting the Hedgehog daemon...", "yellow"))
hedgehog_command = f"nohup {hedgehog_bin} daemon --resthost=0.0.0.0 --restport={rest_port} --netport=40002 --no-seeds --network-keys={generated_public_key} -vvvvvv > hedgehog.log 2>&1 &"
# print(colored(f"Executing: {hedgehog_command}", "yellow"))
execute_command(hedgehog_command)
time.sleep(5)  # Wait for the Hedgehog daemon to start

# Verify if Hedgehog daemon started correctly
try:
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork", verify=False)
    if response.status_code == 200:
        print(colored("Hedgehog daemon started successfully.", "green"))
    else:
        print(colored(f"Failed to start Hedgehog daemon: {response.status_code}", "red"))
except Exception as e:
    print(colored(f"Error connecting to Hedgehog daemon: {e}", "red"))
    sys.exit(1)

# Step 4: Use Ignite to set up the chain
print(colored("Step 4: Setting up the chain using Ignite...", "yellow"))
ignite_command = "nohup ignite chain serve --skip-proto --reset-once --config testing.yml -v > ignite.log 2>&1 &"
# print(colored(f"Executing: {ignite_command}", "yellow"))
execute_command(ignite_command, log_file="ignite.log")

# Step 5: Wait for Ignite chain to be up and running
if not wait_for_ignite_chain(timeout=120):
    sys.exit(colored("Failed to start Ignite chain. Exiting.", "red"))

# Step 6: Generate a new address or account using Ignite
print(colored("Step 6: Generating or fetching a new address...", "yellow"))
# Placeholder for address generation or fetching existing account
address = 'unigrid192yf94yat7h2sfsrawzh694d477ck7gnylwxf3' # Change as needed

# Step 7: Get the current block height
print(colored("Step 7: Getting the current block height...", "yellow"))
current_block_height = get_current_block_height()
print(f"Current Block Height: {current_block_height}")

# Step 8: Generate a random amount for minting
print(colored("Step 8: Generating a random amount for minting...", "yellow"))
mint_amount = random.randint(1000000000, 100000000000)  # Example range for mint amount
print(f"Mint Amount: {mint_amount}")

# Step 9: Submit a mint transaction
print(colored("Step 9: Submitting a mint transaction...", "yellow"))
mint_block_height = int(current_block_height) + 24
mint_url = f"https://127.0.0.1:{rest_port}/gridspork/mint-storage/{address}/{mint_block_height}"
headers = {
    "privateKey": generated_private_key,
    "Content-Type": "application/json"
}
print(f"Executing Mint Transaction: curl -X PUT '{mint_url}' -H 'privateKey: {generated_private_key}' -H 'Content-Type: application/json' -d '{mint_amount}' -k")
mint_response = requests.put(mint_url, headers=headers, data=str(mint_amount), verify=False)
print(f"Mint Response Code: {mint_response.status_code}")

# Verify mint storage after a few seconds
print(colored("Step 10: Verifying mint storage after a short wait...", "yellow"))
time.sleep(5)  # Short wait before checking mint storage
mint_storage_data = verify_mint_storage(rest_port)

# Verify the mint data in Hedgehog
mint_key = f"{address}/{mint_block_height}"
mint_data_correct = mint_storage_data['data']['mints'].get(mint_key) == mint_amount
if mint_data_correct:
    print(colored(f"Mint data verification successful: {mint_key} = {mint_amount}", "green"))
else:
    print(colored(f"Mint data verification failed: expected {mint_key} = {mint_amount}, got {mint_storage_data['data']['mints'].get(mint_key)}", "red"))

# Verify balance after waiting for the block height to be reached
print(colored("Step 11: Waiting for the block height to be reached and verifying the balance...", "yellow"))
time.sleep(30)  # Wait for enough blocks to be mined
balance_url = f"http://127.0.0.1:1317/cosmos/bank/v1beta1/balances/{address}"
print(f"Fetching Balance: {balance_url}")
balance_response = requests.get(balance_url)
balance_data = balance_response.json()
balance = balance_data['balances'][0]['amount'] if 'balances' in balance_data and balance_data['balances'] else '0'

# Step 12: Verify the balance matches the expected amount
expected_balance = mint_amount
if int(balance) == expected_balance:
    print(colored(f"Balance verification successful: {balance} uugd", "green"))
else:
    print(colored(f"Balance verification failed: expected {expected_balance}, got {balance}", "red"))

# Step 13: Get the new block height
print(colored("Step 13: Getting the current block height...", "yellow"))
current_block_height = get_current_block_height()
print(f"Current Block Height: {current_block_height}")

# Step 14: Submit vesting data with the same random amount
print(colored("Step 14: Submitting vesting data...", "yellow"))
new_block_height = int(current_block_height) + 20
start_time = (datetime.now(timezone.utc) + timedelta(minutes=1)).strftime('%Y-%m-%dT%H:%M:%SZ')
cliff = random.randint(0, 12)  # Random cliff between 0 and 12
parts = random.randint(6, 36)  # Random parts between 6 and 36
percent = random.randint(0, 20)  # Random percent between 0 and 20
vesting_data = {
    "amount": int(mint_amount),
    "block": new_block_height,
    "cliff": cliff,
    "duration": "PT5M",
    "parts": parts,
    "percent": percent,
    "start": start_time
}
vesting_url = f"https://127.0.0.1:{rest_port}/gridspork/vesting-storage/{address}"
print(f"Executing Vesting Transaction: curl -X PUT '{vesting_url}' -H 'privateKey: {generated_private_key}' -H 'Content-Type: application/json' -d '{json.dumps(vesting_data)}' -k")
vesting_response = requests.put(vesting_url, headers=headers, json=vesting_data, verify=False)
print(f"Vesting Response Code: {vesting_response.status_code}")
print(f"Vesting Response: {vesting_response.text}")

# Step 15: Wait for the vesting block height to be reached and verify vesting storage
print(colored("Step 15: Waiting for the vesting block height to be reached and verifying vesting storage...", "yellow"))
time.sleep(22)  # Wait for enough blocks to be mined
vesting_storage_data = verify_vesting_storage(rest_port)
# print(f"Vesting Storage Data: {json.dumps(vesting_storage_data, indent=2)}")

# Verify the vesting data in Hedgehog
vesting_data_correct = (
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['amount'] == vesting_data['amount'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['block'] == vesting_data['block'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['cliff'] == vesting_data['cliff'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['duration'] == vesting_data['duration'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['parts'] == vesting_data['parts'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['percent'] == vesting_data['percent'] and
    vesting_storage_data['data']['vestingAddresses'][f"Address(wif={address})"]['start'] == vesting_data['start']
)
if vesting_data_correct:
    print(colored(f"Vesting data verification successful: {json.dumps(vesting_data, indent=2)}", "green"))
else:
    print(colored(f"Vesting data verification failed. Expected: {json.dumps(vesting_data, indent=2)}, Got: {json.dumps(vesting_storage_data['data']['vestingAddresses'][f'Address(wif={address})'], indent=2)}", "red"))

# Step 16: Verify the vesting schedule on the chain
print(colored("Step 16: Verifying the vesting schedule on the chain...", "yellow"))
current_block_height = get_current_block_height()
while int(current_block_height) <= new_block_height:
    print(f"Current Block Height: {current_block_height}, waiting for block height to exceed {new_block_height} for the vesting schedule to appear")
    time.sleep(5)  # Wait before checking again
    current_block_height = get_current_block_height()

vesting_account_data = verify_account_vesting(address)
# print(f"Vesting Account Data: {json.dumps(vesting_account_data, indent=2)}")

# Verify the vesting periods data
vesting_periods = vesting_account_data['account']['vesting_periods']
total_vesting_amount = sum(int(period['amount'][0]['amount']) for period in vesting_periods)
total_parts = len(vesting_periods)

expected_parts = (vesting_data['cliff'] + (vesting_data['parts'] - 1) + 1) if vesting_data['cliff'] > 0 else (vesting_data['parts'] + 1)  # cliff periods + remaining parts + TGE, or parts + TGE if no cliff
if total_vesting_amount == vesting_data['amount'] and total_parts == expected_parts:
    print(colored(f"Vesting periods verification successful: Total Amount = {total_vesting_amount}, Parts = {total_parts}", "green"))
else:
    print(colored(f"Vesting periods verification failed: Expected Amount = {vesting_data['amount']}, Got = {total_vesting_amount}; Expected Parts = {expected_parts}, Got = {total_parts}", "red"))

print(colored("Step 17: Killing any running Hedgehog and Ignite processes after 2 minutes...", "yellow"))
time.sleep(120)
kill_processes("hedgehog-0.0.8")
kill_processes("ignite")
