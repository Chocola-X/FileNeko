"""Check matrix packages and create the combined downloadable checksum file."""

import hashlib
import json
import pathlib
import re
import subprocess
import sys
import tarfile
import tempfile
import zipfile


def verify(folder, archive):
    package = folder / archive
    binary = "fileneko.exe" if archive.startswith("windows-") else "fileneko"
    required = {
        binary, "README.md", "CHANGELOG.md", "LICENSE", "NOTICE",
        "LICENSES/Apache-2.0.txt", "LICENSES/licensing.md",
        *[f"LICENSES/fonts/{font}-OFL.txt" for font in (
            "dancing-script", "fredoka", "jetbrains-mono", "mplus-rounded"
        )],
    }
    if archive.endswith(".zip"):
        with zipfile.ZipFile(package) as opened:
            names = set(opened.namelist())
            content = opened.read(binary)
    else:
        with tarfile.open(package) as opened:
            names = set(opened.getnames())
            member = opened.getmember(binary)
            if not member.isfile() or not member.mode & 0o111:
                raise ValueError(f"{archive}: executable is missing its executable bit")
            content = opened.extractfile(member).read()
    if not required <= names:
        raise ValueError(f"{archive}: missing {sorted(required - names)}")
    if not content:
        raise ValueError(f"{archive}: empty executable")
    with tempfile.TemporaryDirectory(prefix="fileneko-package-") as temporary:
        executable = pathlib.Path(temporary) / binary
        executable.write_bytes(content)
        executable.chmod(0o755)
        subprocess.run(["file", str(executable)], check=True)
        metadata = subprocess.check_output(["go", "version", "-m", str(executable)], text=True)
        if "CGO_ENABLED=0" not in metadata or "-s -w" not in metadata:
            raise ValueError(f"{archive}: expected a static, stripped Go build")
        if archive.startswith("linux-"):
            for option, forbidden in [("-l", "INTERP"), ("-d", "NEEDED")]:
                output = subprocess.check_output(["readelf", option, str(executable)], text=True)
                if forbidden in output:
                    raise ValueError(f"{archive}: unexpected dynamic dependency")
        if archive == "linux-amd64-fileneko.tar.gz":
            subprocess.run([str(executable), "version"], check=True)
    print(f"Verified {archive}")


def checksums(folder, matrix):
    expected = {entry["archive"] for entry in matrix["include"]}
    found = {p.name for p in folder.iterdir() if p.is_file()}
    if not expected or expected != found:
        raise ValueError(f"package set mismatch: missing={sorted(expected-found)}, extra={sorted(found-expected)}")
    lines = []
    for name in sorted(expected):
        if pathlib.Path(name).name != name or not re.fullmatch(r"[a-z0-9-]+\.(tar\.gz|zip)", name):
            raise ValueError(f"invalid package name: {name}")
        with (folder / name).open("rb") as package:
            digest = hashlib.file_digest(package, "sha256").hexdigest()
        lines.append(f"{digest}  {name}\n")
    output = folder / "fileneko_checksums.txt"
    output.write_text("".join(lines))
    subprocess.run(["sha256sum", "--check", output.name], cwd=folder, check=True)
    print(f"Verified all {len(lines)} matrix packages")


if __name__ == "__main__":
    command, folder, value = sys.argv[1:]
    folder = pathlib.Path(folder)
    if command == "verify":
        verify(folder, value)
    elif command == "checksums":
        checksums(folder, json.loads(value))
    else:
        raise ValueError(f"unknown command: {command}")
