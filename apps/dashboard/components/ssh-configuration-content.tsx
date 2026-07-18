"use client";

import { Terminal, Key, Shield, GitBranch, Upload, AlertTriangle, ArrowLeft, ExternalLink, BookOpen, Box } from "lucide-react";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { useEffect, useState } from "react";
import Link from "next/link";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface TocItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

const tocItems: TocItem[] = [
  { id: "prerequisites", label: "Prerequisites", icon: Terminal },
  { id: "generate-ssh-key", label: "Generate SSH Key Pair", icon: Key },
  { id: "add-to-agent", label: "Add Key to ssh-agent", icon: Shield },
  { id: "add-to-hasir", label: "Add Public Key to Hasir", icon: Upload },
  { id: "clone-repository", label: "Clone a Repository", icon: GitBranch },
  { id: "push-changes", label: "Push Changes", icon: Upload },
  { id: "troubleshooting", label: "Troubleshooting", icon: AlertTriangle },
];

function CodeBlock({ language, children }: { language: string; children: string }) {
  return (
    <div className="my-4 overflow-hidden rounded-lg">
      <SyntaxHighlighter
        style={vscDarkPlus}
        language={language}
        PreTag="div"
        customStyle={{
          margin: 0,
          borderRadius: "0.5rem",
          fontSize: "0.875rem",
        }}
      >
        {children.replace(/\n$/, "")}
      </SyntaxHighlighter>
    </div>
  );
}

export function SshConfigurationContent() {
  const [activeSection, setActiveSection] = useState("prerequisites");

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id);
          }
        }
      },
      { rootMargin: "-80px 0px -60% 0px", threshold: 0 },
    );

    for (const item of tocItems) {
      const element = document.getElementById(item.id);
      if (element) {
        observer.observe(element);
      }
    }

    return () => observer.disconnect();
  }, []);

  const handleTocClick = (id: string) => {
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="mx-auto mt-4 flex w-full max-w-6xl items-center gap-4 rounded-md border bg-card px-4 py-2">
        <Link
          href="/"
          className="flex items-center gap-2 text-base font-medium text-foreground"
        >
          <Box className="size-5 text-primary" aria-hidden="true" />
          <span>Hasir</span>
        </Link>
        <div className="flex-1" />
        <Button variant="ghost" size="sm" asChild>
          <Link href="/dashboard">
            <ArrowLeft className="size-4" />
            <span>Back to Dashboard</span>
          </Link>
        </Button>
      </header>

      <div className="h-[calc(100vh-4.5rem)] overflow-hidden px-6 py-6">
        <div className="mx-auto flex h-full w-full max-w-6xl gap-6">
          <Card className="h-full w-64 shrink-0 gap-0 overflow-hidden border border-border py-0">
            <CardHeader className="flex items-center bg-transparent px-6 py-4 shrink-0 border-b border-border">
              <CardTitle className="flex items-center gap-2 text-sm font-mono uppercase tracking-[0.06em] font-medium text-foreground">
                <BookOpen className="size-4" />
                SSH Guide
              </CardTitle>
            </CardHeader>
            <CardContent className="overflow-y-auto py-4">
              <nav className="space-y-1">
                {tocItems.map((item) => {
                  const Icon = item.icon;
                  const isActive = activeSection === item.id;

                  return (
                    <button
                      key={item.id}
                      onClick={() => handleTocClick(item.id)}
                      className={cn(
                        "flex w-full items-center gap-3 rounded-none px-3 py-2.5 text-sm font-mono uppercase tracking-[0.06em] font-medium transition-colors",
                        isActive
                          ? "bg-primary text-primary-foreground"
                          : "text-muted-foreground hover:bg-primary hover:text-primary-foreground"
                      )}
                    >
                      <Icon className="size-4 shrink-0" />
                      <span className="text-left">{item.label}</span>
                    </button>
                  );
                })}
              </nav>
            </CardContent>
          </Card>

          <main className="flex-1 space-y-6 overflow-y-auto">
            <div className="flex items-center gap-4">
              <div>
                <h1 className="text-2xl font-semibold">SSH Configuration Guide</h1>
                <p className="text-muted-foreground mt-1 text-sm">
                  Set up SSH keys for secure access to your Hasir repositories
                </p>
              </div>
            </div>

            <div className="space-y-10 pb-12">
              {/* Prerequisites */}
              <section id="prerequisites">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <Terminal className="size-5 text-primary" />
                  Prerequisites
                </h2>
                <Separator className="mb-4" />
                <p className="text-muted-foreground mb-4 text-sm leading-relaxed">
                  Before configuring SSH access, ensure you have the following:
                </p>
                <ul className="text-muted-foreground mb-4 list-inside list-disc space-y-2 text-sm">
                  <li>
                    A <strong className="text-foreground">terminal application</strong> (Terminal on macOS, PowerShell or Git Bash on Windows, any terminal emulator on Linux)
                  </li>
                  <li>
                    An <strong className="text-foreground">SSH client</strong> installed (included by default on macOS and most Linux distributions; on Windows, install{" "}
                    <a
                      href="https://git-scm.com"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      Git for Windows
                      <ExternalLink className="ml-1 inline size-3" />
                    </a>{" "}
                    or enable the built-in OpenSSH client)
                  </li>
                  <li>
                    A <strong className="text-foreground">Hasir account</strong> with access to at least one repository
                  </li>
                  <li>
                    <strong className="text-foreground">Git</strong> installed on your system
                  </li>
                </ul>
                <p className="text-muted-foreground text-sm leading-relaxed">
                  Verify your SSH installation by running:
                </p>
                <CodeBlock language="bash">{`ssh -V
# Example output: OpenSSH_9.6p1, LibreSSL 3.3.6`}</CodeBlock>
              </section>

              {/* Generate SSH Key Pair */}
              <section id="generate-ssh-key">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <Key className="size-5 text-primary" />
                  Generate SSH Key Pair
                </h2>
                <Separator className="mb-4" />
                <p className="text-muted-foreground mb-4 text-sm leading-relaxed">
                  Generate a new SSH key pair to authenticate with Hasir. We recommend <strong className="text-foreground">Ed25519</strong> for its security and performance.
                </p>

                <h3 className="mb-2 text-base font-semibold">Recommended: Ed25519</h3>
                <CodeBlock language="bash">{`ssh-keygen -t ed25519 -C "your_email@example.com"`}</CodeBlock>

                <h3 className="mb-2 mt-6 text-base font-semibold">Fallback: RSA (4096-bit)</h3>
                <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                  If your system does not support Ed25519, use RSA with a 4096-bit key:
                </p>
                <CodeBlock language="bash">{`ssh-keygen -t rsa -b 4096 -C "your_email@example.com"`}</CodeBlock>

                <Alert className="mt-4">
                  <Shield className="size-4" />
                  <AlertTitle>Key generation prompts</AlertTitle>
                  <AlertDescription>
                    <p>When prompted:</p>
                    <ul className="mt-1 list-inside list-disc space-y-1">
                      <li><strong>File location</strong> — Press Enter to accept the default (<code className="rounded bg-muted px-1.5 py-0.5 text-xs">~/.ssh/id_ed25519</code>)</li>
                      <li><strong>Passphrase</strong> — Enter a strong passphrase for additional security (recommended) or press Enter for no passphrase</li>
                    </ul>
                  </AlertDescription>
                </Alert>

                <p className="text-muted-foreground mt-4 text-sm leading-relaxed">
                  This creates two files:
                </p>
                <ul className="text-muted-foreground mt-2 list-inside list-disc space-y-1 text-sm">
                  <li><code className="rounded bg-muted px-1.5 py-0.5 text-xs">~/.ssh/id_ed25519</code> — Your <strong className="text-foreground">private key</strong> (never share this)</li>
                  <li><code className="rounded bg-muted px-1.5 py-0.5 text-xs">~/.ssh/id_ed25519.pub</code> — Your <strong className="text-foreground">public key</strong> (this is what you add to Hasir)</li>
                </ul>
              </section>

              {/* Add SSH Key to ssh-agent */}
              <section id="add-to-agent">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <Shield className="size-5 text-primary" />
                  Add SSH Key to ssh-agent
                </h2>
                <Separator className="mb-4" />
                <p className="text-muted-foreground mb-4 text-sm leading-relaxed">
                  The SSH agent manages your keys so you do not have to enter your passphrase every time you connect.
                </p>

                <h3 className="mb-2 text-base font-semibold">macOS</h3>
                <CodeBlock language="bash">{`# Start the ssh-agent
eval "$(ssh-agent -s)"

# Add your key to the macOS Keychain
ssh-add --apple-use-keychain ~/.ssh/id_ed25519`}</CodeBlock>
                <p className="text-muted-foreground mb-4 mt-2 text-sm leading-relaxed">
                  To persist across reboots, add this to <code className="rounded bg-muted px-1.5 py-0.5 text-xs">~/.ssh/config</code>:
                </p>
                <CodeBlock language="bash">{`Host *
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/id_ed25519`}</CodeBlock>

                <h3 className="mb-2 mt-6 text-base font-semibold">Linux</h3>
                <CodeBlock language="bash">{`# Start the ssh-agent
eval "$(ssh-agent -s)"

# Add your key
ssh-add ~/.ssh/id_ed25519`}</CodeBlock>

                <h3 className="mb-2 mt-6 text-base font-semibold">Windows (Git Bash)</h3>
                <CodeBlock language="bash">{`# Start the ssh-agent
eval "$(ssh-agent -s)"

# Add your key
ssh-add ~/.ssh/id_ed25519`}</CodeBlock>

                <h3 className="mb-2 mt-6 text-base font-semibold">Windows (PowerShell)</h3>
                <CodeBlock language="powershell">{`# Ensure the ssh-agent service is running
Get-Service ssh-agent | Set-Service -StartupType Automatic
Start-Service ssh-agent

# Add your key
ssh-add $env:USERPROFILE\\.ssh\\id_ed25519`}</CodeBlock>
              </section>

              {/* Add Public Key to Hasir */}
              <section id="add-to-hasir">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <Upload className="size-5 text-primary" />
                  Add Public Key to Hasir
                </h2>
                <Separator className="mb-4" />

                <div className="space-y-4">
                  <div>
                    <h3 className="mb-2 text-base font-semibold">Step 1: Copy your public key</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      Print the contents of your public key file:
                    </p>
                    <CodeBlock language="bash">{`# macOS
pbcopy < ~/.ssh/id_ed25519.pub

# Linux (requires xclip)
xclip -selection clipboard < ~/.ssh/id_ed25519.pub

# Windows (PowerShell)
Get-Content $env:USERPROFILE\\.ssh\\id_ed25519.pub | Set-Clipboard

# Or just print it and copy manually
cat ~/.ssh/id_ed25519.pub`}</CodeBlock>
                    <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                      The public key looks something like:
                    </p>
                    <CodeBlock language="text">{`ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... your_email@example.com`}</CodeBlock>
                  </div>

                  <div>
                    <h3 className="mb-2 text-base font-semibold">Step 2: Add the key to your Hasir profile</h3>
                    <ol className="text-muted-foreground mb-4 list-inside list-decimal space-y-2 text-sm">
                      <li>Navigate to your <strong className="text-foreground">Profile</strong> page</li>
                      <li>Find the <strong className="text-foreground">SSH Keys</strong> section</li>
                      <li>Click <strong className="text-foreground">Add SSH Key</strong></li>
                      <li>Paste your public key into the key field</li>
                      <li>Give the key a descriptive name (e.g., &quot;Work Laptop&quot;, &quot;Home Desktop&quot;)</li>
                      <li>Click <strong className="text-foreground">Save</strong></li>
                    </ol>
                    <Button variant="outline" asChild>
                      <Link href="/profile">
                        Go to Profile Settings
                        <ExternalLink className="size-4" />
                      </Link>
                    </Button>
                  </div>
                </div>
              </section>

              {/* Clone a Repository */}
              <section id="clone-repository">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <GitBranch className="size-5 text-primary" />
                  Clone a Repository
                </h2>
                <Separator className="mb-4" />
                <p className="text-muted-foreground mb-4 text-sm leading-relaxed">
                  Once your SSH key is added, you can clone repositories using the SSH URL. The format is:
                </p>
                <CodeBlock language="text">{`ssh://git@<host>:<port>/<repositoryId>.git`}</CodeBlock>
                <p className="text-muted-foreground mb-2 mt-4 text-sm leading-relaxed">
                  For example, to clone a repository:
                </p>
                <CodeBlock language="bash">{`git clone ssh://git@ssh.hasir.dev:2222/my-repository-id.git`}</CodeBlock>

                <Alert className="mt-4">
                  <BookOpen className="size-4" />
                  <AlertTitle>Finding the clone URL</AlertTitle>
                  <AlertDescription>
                    You can find the exact SSH clone URL on the repository page in the Hasir dashboard. Look for the <strong>Clone</strong> button in the top-right corner of the repository view.
                  </AlertDescription>
                </Alert>

                <p className="text-muted-foreground mt-4 text-sm leading-relaxed">
                  After cloning, navigate into the repository:
                </p>
                <CodeBlock language="bash">{`cd my-repository-id`}</CodeBlock>
              </section>

              {/* Push Changes */}
              <section id="push-changes">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <Upload className="size-5 text-primary" />
                  Push Changes
                </h2>
                <Separator className="mb-4" />
                <p className="text-muted-foreground mb-4 text-sm leading-relaxed">
                  After making changes to your protobuf schema files, push them to Hasir using the standard Git workflow:
                </p>
                <CodeBlock language="bash">{`# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add user service definition"

# Push to the remote repository
git push origin main`}</CodeBlock>

                <Alert className="mt-4">
                  <Terminal className="size-4" />
                  <AlertTitle>First push</AlertTitle>
                  <AlertDescription>
                    If you are pushing to a newly created repository, you may need to set the upstream branch:
                    <CodeBlock language="bash">{`git push -u origin main`}</CodeBlock>
                  </AlertDescription>
                </Alert>
              </section>

              {/* Troubleshooting */}
              <section id="troubleshooting">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                  <AlertTriangle className="size-5 text-primary" />
                  Troubleshooting
                </h2>
                <Separator className="mb-4" />

                <div className="space-y-6">
                  <div>
                    <h3 className="mb-2 text-base font-semibold">Permission denied (publickey)</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      This error means Hasir could not authenticate you. Common causes:
                    </p>
                    <ul className="text-muted-foreground mb-2 list-inside list-disc space-y-1 text-sm">
                      <li>Your SSH key is not added to your Hasir profile</li>
                      <li>You are using a different key than the one registered</li>
                      <li>The ssh-agent does not have your key loaded</li>
                    </ul>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      Verify your key is loaded:
                    </p>
                    <CodeBlock language="bash">{`# List keys loaded in the agent
ssh-add -l

# If empty, add your key
ssh-add ~/.ssh/id_ed25519`}</CodeBlock>
                  </div>

                  <div>
                    <h3 className="mb-2 text-base font-semibold">Connection refused</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      If you see <code className="rounded bg-muted px-1.5 py-0.5 text-xs">ssh: connect to host ... port ...: Connection refused</code>, check:
                    </p>
                    <ul className="text-muted-foreground mb-2 list-inside list-disc space-y-1 text-sm">
                      <li>The hostname and port in your clone URL are correct</li>
                      <li>Your network allows outbound SSH connections on the specified port</li>
                      <li>If you are behind a corporate firewall, you may need to configure an SSH proxy</li>
                    </ul>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      Test the connection with verbose output:
                    </p>
                    <CodeBlock language="bash">{`ssh -vT git@ssh.hasir.dev -p 2222`}</CodeBlock>
                  </div>

                  <div>
                    <h3 className="mb-2 text-base font-semibold">Key not accepted</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      If your key is not being accepted:
                    </p>
                    <ul className="text-muted-foreground mb-2 list-inside list-disc space-y-1 text-sm">
                      <li>Ensure you copied the <strong className="text-foreground">public</strong> key (ending in <code className="rounded bg-muted px-1.5 py-0.5 text-xs">.pub</code>), not the private key</li>
                      <li>Make sure there are no extra newlines or spaces when pasting</li>
                      <li>Check your key file permissions (private key should be <code className="rounded bg-muted px-1.5 py-0.5 text-xs">600</code>)</li>
                    </ul>
                    <CodeBlock language="bash">{`# Fix private key permissions
chmod 600 ~/.ssh/id_ed25519
chmod 644 ~/.ssh/id_ed25519.pub`}</CodeBlock>
                  </div>

                  <div>
                    <h3 className="mb-2 text-base font-semibold">Multiple SSH keys</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      If you use different SSH keys for different services, configure the <code className="rounded bg-muted px-1.5 py-0.5 text-xs">~/.ssh/config</code> file to specify which key to use for Hasir:
                    </p>
                    <CodeBlock language="bash">{`# ~/.ssh/config

# Hasir
Host ssh.hasir.dev
  HostName ssh.hasir.dev
  Port 2222
  User git
  IdentityFile ~/.ssh/id_ed25519_hasir
  IdentitiesOnly yes

# GitHub (example)
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_github
  IdentitiesOnly yes`}</CodeBlock>
                    <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                      With this configuration, SSH will automatically use the correct key when connecting to each host.
                    </p>
                  </div>

                  <div>
                    <h3 className="mb-2 text-base font-semibold">Debug SSH connection</h3>
                    <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
                      For detailed connection debugging, use verbose mode:
                    </p>
                    <CodeBlock language="bash">{`# Increasing levels of verbosity
ssh -v  git@ssh.hasir.dev -p 2222  # verbose
ssh -vv git@ssh.hasir.dev -p 2222  # more verbose
ssh -vvv git@ssh.hasir.dev -p 2222 # maximum verbosity`}</CodeBlock>
                  </div>
                </div>
              </section>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
