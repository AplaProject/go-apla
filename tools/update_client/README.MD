# Apla update client
### Available commands (use `./update-client command-name --help` for detailed command params):
* add-binary
* generate-keys
* get-binary
* remove-binary

### Examples:
* ./update_client add-binary --binary-path=./resources/test-binary --login=admin --password=admin --server=http://localhost:12345 --start-block=1 --version=linux_amd64_1.2
* ./update_client get-binary --server=http://localhost:12345 --version=linux_amd64_1.2 --binary-path=./resources
* ./update_client remove-binary --login=admin --password=admin --server=http://localhost:12345 --version=linux_amd64_1.0
* ./update_client versions --server=http://localhost:12345 --version=windows_i386_
