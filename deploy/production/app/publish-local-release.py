import argparse
import getpass
import os
import posixpath
import sys
import time

import paramiko

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def parse_args():
    parser = argparse.ArgumentParser(
        description="Upload a locally built new-api release and restart the server container."
    )
    parser.add_argument("--host", default="43.173.102.239")
    parser.add_argument("--port", type=int, default=22)
    parser.add_argument("--user", default="root")
    parser.add_argument("--remote-app-dir", default="/srv/new-api/app")
    parser.add_argument(
        "--password-env",
        default="NEW_API_SSH_PASSWORD",
        help="Environment variable that stores the SSH password.",
    )
    parser.add_argument(
        "--local-compose",
        default=os.path.join(os.path.dirname(__file__), "docker-compose.yml"),
    )
    parser.add_argument(
        "--local-binary",
        default=os.path.join(os.path.dirname(__file__), "release", "new-api"),
    )
    return parser.parse_args()


def wait_for_healthy(client):
    check = "docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}' new-api-app"
    for attempt in range(1, 31):
        _, stdout, _ = client.exec_command(check, timeout=30)
        status = stdout.read().decode("utf-8", errors="replace").strip()
        print(f"health_check_attempt={attempt} status={status}")
        if status == "healthy":
            return
        time.sleep(2)
    raise RuntimeError("container did not become healthy within 60 seconds")


def main():
    args = parse_args()
    password = os.getenv(args.password_env) or getpass.getpass(
        f"SSH password for {args.user}@{args.host}: "
    )

    local_compose = os.path.abspath(args.local_compose)
    local_binary = os.path.abspath(args.local_binary)
    if not os.path.exists(local_compose):
        raise FileNotFoundError(f"compose file not found: {local_compose}")
    if not os.path.exists(local_binary):
        raise FileNotFoundError(f"release binary not found: {local_binary}")

    remote_compose = posixpath.join(
        args.remote_app_dir, "deploy/production/app/docker-compose.yml"
    )
    remote_release_dir = posixpath.join(
        args.remote_app_dir, "deploy/production/app/release"
    )
    remote_binary = posixpath.join(remote_release_dir, "new-api")
    remote_binary_tmp = posixpath.join(remote_release_dir, "new-api.uploading")

    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(
        args.host,
        port=args.port,
        username=args.user,
        password=password,
        timeout=20,
    )

    try:
        for remote_dir in (
            posixpath.dirname(remote_compose),
            remote_release_dir,
        ):
            _, stdout, stderr = client.exec_command(f"mkdir -p {remote_dir}", timeout=30)
            exit_code = stdout.channel.recv_exit_status()
            if exit_code != 0:
                raise RuntimeError(stderr.read().decode("utf-8", errors="replace"))

        sftp = client.open_sftp()
        try:
            sftp.put(local_compose, remote_compose)
            sftp.put(local_binary, remote_binary_tmp)
            sftp.chmod(remote_binary_tmp, 0o755)
        finally:
            sftp.close()

        swap_cmd = (
            f"mv {remote_binary_tmp} {remote_binary} && "
            f"chmod 755 {remote_binary}"
        )
        _, stdout, stderr = client.exec_command(swap_cmd, timeout=60)
        exit_code = stdout.channel.recv_exit_status()
        if exit_code != 0:
            raise RuntimeError(stderr.read().decode("utf-8", errors="replace"))

        deploy_cmd = (
            f"cd {args.remote_app_dir} && "
            "docker compose --env-file deploy/production/app/.env "
            "-f deploy/production/app/docker-compose.yml up -d"
        )
        _, stdout, stderr = client.exec_command(deploy_cmd, timeout=1800)
        exit_code = stdout.channel.recv_exit_status()
        sys.stdout.write(stdout.read().decode("utf-8", errors="replace"))
        sys.stderr.write(stderr.read().decode("utf-8", errors="replace"))
        if exit_code != 0:
            raise RuntimeError("docker compose up failed")

        wait_for_healthy(client)

        verify_cmd = (
            f"cd {args.remote_app_dir} && "
            "docker compose --env-file deploy/production/app/.env "
            "-f deploy/production/app/docker-compose.yml ps && "
            "echo '=== api status ===' && "
            "curl -fsS http://127.0.0.1:3000/api/status && "
            "echo && echo '=== logs tail ===' && "
            "docker compose --env-file deploy/production/app/.env "
            "-f deploy/production/app/docker-compose.yml logs --tail=80"
        )
        _, stdout, stderr = client.exec_command(verify_cmd, timeout=600)
        exit_code = stdout.channel.recv_exit_status()
        sys.stdout.write(stdout.read().decode("utf-8", errors="replace"))
        sys.stderr.write(stderr.read().decode("utf-8", errors="replace"))
        if exit_code != 0:
            raise RuntimeError("post-deploy verification failed")
    finally:
        client.close()


if __name__ == "__main__":
    main()
