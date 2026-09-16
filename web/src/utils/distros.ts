/**
 * Copyright 2023 sigma
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const distroAliases: Record<string, string> = {
  "alma linux": "alma",
  almalinux: "alma",
  "alpine linux": "alpine",
  alpinelinux: "alpine",
  "arch linux": "arch",
  archlinux: "arch",
  "centos stream": "centos",
  "linux mint": "mint",
  opensuse: "suse",
  "opensuse leap": "suse",
  "opensuse tumbleweed": "suse",
  "pop os": "pop",
  "pop! os": "pop",
  "pop!_os": "pop",
  "raspberry pi os": "raspios",
  raspbian: "raspios",
  "red hat": "redhat",
  "red hat enterprise linux": "redhat",
  rhel: "redhat",
  "suse linux": "suse",
  "suse linux enterprise": "suse",
  "void linux": "void",
  "zorin os": "zorin",
};

const distroIcons: Record<string, string> = {
  alma: "alma.svg",
  alpine: "alpine.svg",
  antix: "linux.svg",
  arch: "arch.svg",
  bunsenlabs: "linux.svg",
  centos: "centos.svg",
  clear: "linux.svg",
  debian: "debian.svg",
  deepin: "deepin.svg",
  elementary: "elementary.svg",
  endless: "linux.svg",
  fedora: "fedora.svg",
  gentoo: "gentoo.svg",
  kali: "kali.svg",
  kaos: "linux.svg",
  kubuntu: "kubuntu.svg",
  linux: "linux.svg",
  lite: "linux.svg",
  mabox: "linux.svg",
  manjaro: "manjaro.svg",
  mint: "mint.svg",
  nixos: "nixos.svg",
  pop: "pop.svg",
  raspios: "raspios.svg",
  redhat: "redhat.svg",
  suse: "suse.svg",
  ubuntu: "ubuntu.svg",
  void: "void.svg",
  windows: "windows.svg",
  xubuntu: "xubuntu.svg",
  zorin: "zorin.svg",
};

const distroDisplayNames: Record<string, string> = {
  alma: "AlmaLinux",
  alpine: "Alpine",
  antix: "antiX",
  arch: "Arch Linux",
  bunsenlabs: "BunsenLabs",
  centos: "CentOS",
  clear: "Clear Linux",
  debian: "Debian",
  deepin: "Deepin",
  elementary: "elementary OS",
  endless: "Endless OS",
  fedora: "Fedora",
  gentoo: "Gentoo",
  kali: "Kali Linux",
  kaos: "KaOS",
  kubuntu: "Kubuntu",
  linux: "Linux",
  lite: "Linux Lite",
  mabox: "Mabox Linux",
  manjaro: "Manjaro",
  mint: "Linux Mint",
  nixos: "NixOS",
  pop: "Pop!_OS",
  raspios: "Raspberry Pi OS",
  redhat: "Red Hat Enterprise Linux",
  suse: "SUSE",
  ubuntu: "Ubuntu",
  void: "Void Linux",
  windows: "Windows",
  xubuntu: "Xubuntu",
  zorin: "Zorin OS",
};

function normalizeDistroName(name: string) {
  const normalized = name.trim().toLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");
  const compact = normalized.replace(/\s+/g, "");
  return distroAliases[normalized] ?? distroAliases[compact] ?? compact;
}

export default function (name: string) {
  const distro = normalizeDistroName(name);
  return distroIcons[distro] ?? "";
}

export function distroName(name: string) {
  const distro = normalizeDistroName(name);
  return distroDisplayNames[distro] ?? "";
}
