#!/usr/bin/env bash

script_name="evilgophish setup"

# Define hardcoded certificate path base
certs_base_path="/etc/letsencrypt/live"

function check_privs () {
    if [[ "$(whoami)" != "root" ]]; then
        print_error "You need root privileges to run this script."
        exit 1
    fi
}

function print_good () {
    echo -e "[${script_name}] \x1B[01;32m[+]\x1B[0m $1"
}

function print_error () {
    echo -e "[${script_name}] \x1B[01;31m[-]\x1B[0m $1"
}

function print_warning () {
    echo -e "[${script_name}] \x1B[01;33m[-]\x1B[0m $1"
}

function print_info () {
    echo -e "[${script_name}] \x1B[01;34m[*]\x1B[0m $1"
}

# Main installation steps
function main () {
    check_privs
    install_depends
    generate_certs "${root_domain}" "${evilginx3_subs}"
    setup_apache
    setup_gophish
    setup_evilginx3
    print_good "Installation complete! When ready start apache with: systemctl restart apache2"
    print_info "It is recommended to run all servers inside a tmux session to avoid losing them over SSH!"
}

# Install needed dependencies
function install_depends () {
    print_info "Installing dependencies with apt"
    apt-get update
    apt-get install apache2 build-essential letsencrypt certbot wget git net-tools tmux openssl jq -y
    print_good "Installed dependencies with apt!"
    print_info "Installing Go from source"
    v=$(curl -s https://go.dev/dl/?mode=json | jq -r '.[0].version')
    wget https://go.dev/dl/"${v}".linux-amd64.tar.gz
    tar -C /usr/local -xzf "${v}".linux-amd64.tar.gz
    ln -sf /usr/local/go/bin/go /usr/bin/go
    rm "${v}".linux-amd64.tar.gz
    print_good "Installed Go from source!"
}

# Generate certificates with Certbot
function generate_certs () {
    local root_domain=$1
    local subdomains=$2
    certs_path="${certs_base_path}/${root_domain}/" # Path where certbot stores certs

    print_info "Generating certificates for ${root_domain} and its subdomains"
    chmod +x /tmp/auth-hook.sh && chmod +x /tmp/cleanup-hook.sh

    # Prepare the domain arguments for certbot
    domain_args="-d ${root_domain}"
    for subdomain in ${subdomains}; do
        full_domain="${subdomain}.${root_domain}"
        domain_args="${domain_args} -d ${full_domain}"
    done

    certbot certonly --manual --preferred-challenges=dns --manual-auth-hook /tmp/auth-hook.sh --manual-cleanup-hook /tmp/cleanup-hook.sh --email "admin@${root_domain}" --server https://acme-v02.api.letsencrypt.org/directory --agree-tos ${domain_args} --no-eff-email --manual-public-ip-logging-ok

    # Validate cert generation
    if [[ -d "${certs_path}" ]]; then
        print_good "Certificates generated at ${certs_path}"
    else
        print_error "Failed to generate certificates?"
    fi
}

if [[ $# -ne 7 ]]; then
    print_error "Missing Parameters. Usage: $0 <root domain> <subdomains> <root domain bool> <redirect url> <feed bool> <rid replacement> <blacklist bool>"
    exit 2
fi

# Set variables from parameters
root_domain="${1}"
evilginx3_subs="${2}"
e_root_bool="${3}"
redirect_url="${4}"
feed_bool="${5}"
rid_replacement="${6}"
evilginx_dir=$HOME/.evilginx
bl_bool="${7}"

main
