"use client";

import type { FileTreeNode } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { RegistryService } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { NodeType } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { useCallback, useContext, useEffect, useState } from "react";
import { Code, ConnectError } from "@connectrpc/connect";
import { Files, FileText } from "lucide-react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  TreeWithLazyLoad,
  type ExtendedTreeViewElement,
} from "@/components/ui/file-tree";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { RepositoryContext } from "@/lib/repository-context";
import { Skeleton } from "@/components/ui/skeleton";
import { useClient } from "@/lib/use-client";

function transformFileTreeNode(
  node: FileTreeNode,
  markDirectoriesAsUnloaded = false
): ExtendedTreeViewElement {
  const isDirectory = node.type === NodeType.DIRECTORY;

  if (isDirectory && markDirectoriesAsUnloaded) {
    return {
      id: node.path,
      name: node.name,
      children: [],
      isLoaded: false,
      isLoading: false,
    };
  }

  return {
    id: node.path,
    name: node.name,
    children: isDirectory
      ? node.children.map((child) =>
          transformFileTreeNode(child, markDirectoriesAsUnloaded)
        )
      : undefined,
    isLoaded: isDirectory ? true : undefined,
    isLoading: false,
  };
}

function FileTreeSkeleton() {
  return (
    <div className="space-y-2 p-2">
      <Skeleton className="h-6 w-3/4" />
      <div className="ml-4 space-y-2">
        <Skeleton className="h-6 w-5/6" />
        <Skeleton className="h-6 w-4/6" />
        <Skeleton className="h-6 w-5/6" />
      </div>
      <Skeleton className="h-6 w-2/3" />
      <div className="ml-4 space-y-2">
        <Skeleton className="h-6 w-4/5" />
        <Skeleton className="h-6 w-3/5" />
      </div>
    </div>
  );
}

function getLanguageFromPath(path: string): string {
  const extension = path.split(".").pop()?.toLowerCase() || "";
  const languageMap: Record<string, string> = {
    js: "javascript",
    jsx: "jsx",
    ts: "typescript",
    tsx: "tsx",
    json: "json",
    md: "markdown",
    proto: "protobuf",
    py: "python",
    go: "go",
    rs: "rust",
    java: "java",
    cpp: "cpp",
    c: "c",
    sh: "bash",
    yaml: "yaml",
    yml: "yaml",
    toml: "toml",
    xml: "xml",
    html: "html",
    css: "css",
    scss: "scss",
    sql: "sql",
    graphql: "graphql",
    dockerfile: "docker",
  };
  return languageMap[extension] || "text";
}

function isTextFile(mimeType: string): boolean {
  if (mimeType && mimeType.startsWith("text/")) return true;

  const textLikeTypes = new Set([
    "application/json",
    "application/xml",
    "application/javascript",
    "application/x-sh",
    "application/yaml",
    "application/x-yaml",
    "text/x-yaml",
  ]);

  if (mimeType && textLikeTypes.has(mimeType)) return true;

  return false;
}

function formatFileSize(bytes: bigint): string {
  const sizes = ["B", "KB", "MB", "GB"];
  if (bytes === BigInt(0)) return "0 B";
  const i = Math.floor(Math.log(Number(bytes)) / Math.log(1024));
  return `${(Number(bytes) / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
}

interface FilePreview {
  content: string;
  mimeType: string;
  size: bigint;
}

export default function RepositoryFilesContent() {
  const [selectedFile, setSelectedFile] = useState<string | undefined>();
  const [fileTree, setFileTree] = useState<ExtendedTreeViewElement[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [filePreview, setFilePreview] = useState<FilePreview | null>(null);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);
  const [previewError, setPreviewError] = useState<Error | null>(null);

  const context = useContext(RepositoryContext);
  const client = useClient(RegistryService);

  if (!context) {
    throw new Error(
      "RepositoryFilesContent must be used within RepositoryContext"
    );
  }

  const { repository } = context;

  const loadRootTree = useCallback(async () => {
    if (!repository?.id) return;

    try {
      setIsLoading(true);
      setError(null);
      const response = await client.getFileTree({ id: repository.id });
      const rootNodes = response.nodes.map((node) =>
        transformFileTreeNode(node, true)
      );
      setFileTree(rootNodes);
    } catch (err) {
      setError(
        err instanceof Error ? err : new Error("Failed to load file tree")
      );
    } finally {
      setIsLoading(false);
    }
  }, [repository?.id, client]);

  const updateNodeLoading = useCallback(
    (
      nodes: ExtendedTreeViewElement[],
      targetPath: string,
      loading: boolean
    ): ExtendedTreeViewElement[] => {
      return nodes.map((node) => {
        if (node.id === targetPath) {
          return { ...node, isLoading: loading };
        }
        if (node.children) {
          return {
            ...node,
            children: updateNodeLoading(node.children, targetPath, loading),
          };
        }
        return node;
      });
    },
    []
  );

  const updateNodeChildren = useCallback(
    (
      nodes: ExtendedTreeViewElement[],
      targetPath: string,
      newChildren: ExtendedTreeViewElement[]
    ): ExtendedTreeViewElement[] => {
      return nodes.map((node) => {
        if (node.id === targetPath) {
          return {
            ...node,
            children: newChildren,
            isLoaded: true,
            isLoading: false,
          };
        }
        if (node.children) {
          return {
            ...node,
            children: updateNodeChildren(
              node.children,
              targetPath,
              newChildren
            ),
          };
        }
        return node;
      });
    },
    []
  );

  const findNodeInTree = useCallback(
    (
      nodes: ExtendedTreeViewElement[],
      id: string
    ): ExtendedTreeViewElement | null => {
      for (const node of nodes) {
        if (node.id === id) {
          return node;
        }
        if (node.children && node.children.length > 0) {
          const found = findNodeInTree(node.children, id);
          if (found) return found;
        }
      }
      return null;
    },
    []
  );

  const loadFolderContents = useCallback(
    async (folderPath: string) => {
      if (!repository?.id) return;

      setFileTree((currentTree) => {
        const folderNode = findNodeInTree(currentTree, folderPath);
        if (!folderNode || folderNode.isLoaded || folderNode.isLoading) {
          return currentTree;
        }
        return updateNodeLoading(currentTree, folderPath, true);
      });

      try {
        const response = await client.getFileTree({
          id: repository.id,
          path: folderPath,
        });

        setFileTree((prev) =>
          updateNodeChildren(
            prev,
            folderPath,
            response.nodes.map((node) => transformFileTreeNode(node, true))
          )
        );
      } catch {
        setFileTree((prev) => updateNodeLoading(prev, folderPath, false));
      }
    },
    [
      repository?.id,
      client,
      updateNodeLoading,
      updateNodeChildren,
      findNodeInTree,
    ]
  );

  const loadFilePreview = useCallback(
    async (filePath: string) => {
      if (!repository?.id) return;

      setIsLoadingPreview(true);
      setPreviewError(null);
      setFilePreview(null);

      try {
        const response = await client.getFilePreview({
          id: repository.id,
          path: filePath,
        });

        setFilePreview({
          content: response.content,
          mimeType: response.mimeType,
          size: response.size,
        });
      } catch (err) {
        setPreviewError(
          err instanceof Error ? err : new Error("Failed to load file preview")
        );
      } finally {
        setIsLoadingPreview(false);
      }
    },
    [repository?.id, client]
  );

  useEffect(() => {
    loadRootTree();
  }, [loadRootTree]);

  useEffect(() => {
    if (selectedFile) {
      loadFilePreview(selectedFile);
    } else {
      setFilePreview(null);
      setPreviewError(null);
    }
  }, [selectedFile, loadFilePreview]);

  const isNotFound =
    error instanceof ConnectError && error.code === Code.NotFound;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Files className="size-5 text-muted-foreground" />
            <CardTitle>Files</CardTitle>
          </div>
          <CardDescription className="mt-1.5">
            Browse and manage repository files and schemas
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <div className="lg:col-span-1">
              <div className="rounded-lg border border-border">
                <div className="border-border border-b bg-muted/50 p-3">
                  <h3 className="text-sm font-medium">File Tree</h3>
                </div>
                <div className="h-[500px] overflow-auto">
                  {isLoading ? (
                    <FileTreeSkeleton />
                  ) : error && !isNotFound ? (
                    <div className="p-4">
                      <Alert variant="destructive">
                        <AlertTitle>Error loading file tree</AlertTitle>
                        <AlertDescription>
                          {error instanceof ConnectError
                            ? error.message
                            : "Failed to load file tree. Please try again later."}
                        </AlertDescription>
                      </Alert>
                    </div>
                  ) : fileTree.length === 0 || isNotFound ? (
                    <div className="text-muted-foreground flex h-full items-center justify-center px-4 text-center">
                      <p className="text-sm">
                        No files found in this repository.
                      </p>
                    </div>
                  ) : (
                    <div className="p-2">
                      <TreeWithLazyLoad
                        fileTree={fileTree}
                        selectedFile={selectedFile}
                        onFolderExpand={loadFolderContents}
                        onFileSelect={setSelectedFile}
                      />
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="lg:col-span-2">
              <div className="rounded-lg border border-border">
                <div className="border-border border-b bg-muted/50 p-3">
                  <h3 className="text-sm font-medium">
                    {selectedFile ? "File Preview" : "Select a file"}
                  </h3>
                </div>
                <div className="h-[500px] overflow-auto p-4">
                  {selectedFile ? (
                    <div className="space-y-4">
                      {isLoadingPreview ? (
                        <div className="space-y-2">
                          <Skeleton className="h-6 w-1/2" />
                          <Skeleton className="h-[300px] w-full" />
                        </div>
                      ) : previewError ? (
                        <Alert variant="destructive">
                          <AlertTitle>Error loading file preview</AlertTitle>
                          <AlertDescription>
                            {previewError instanceof ConnectError
                              ? previewError.message
                              : "Failed to load file preview. Please try again."}
                          </AlertDescription>
                        </Alert>
                      ) : filePreview ? (
                        <>
                          <div className="mb-4 space-y-1">
                            <div className="flex items-center gap-2">
                              <FileText className="size-4 text-muted-foreground" />
                              <h4 className="text-sm font-medium">
                                {selectedFile.split("/").pop()}
                              </h4>
                            </div>
                            <p className="text-muted-foreground text-xs">
                              {formatFileSize(filePreview.size)} •{" "}
                              {filePreview.mimeType}
                            </p>
                          </div>
                          <div className="rounded-lg border border-border overflow-hidden">
                            {isTextFile(filePreview.mimeType) ? (
                              <SyntaxHighlighter
                                language={getLanguageFromPath(selectedFile)}
                                style={vscDarkPlus}
                                customStyle={{
                                  margin: 0,
                                  borderRadius: 0,
                                  fontSize: "0.875rem",
                                }}
                                showLineNumbers
                              >
                                {filePreview.content}
                              </SyntaxHighlighter>
                            ) : (
                              <div className="bg-muted/30 p-8 text-center">
                                <FileText className="mx-auto mb-4 size-12 text-muted-foreground" />
                                <p className="text-muted-foreground text-sm">
                                  Preview not available for this file type
                                </p>
                              </div>
                            )}
                          </div>
                        </>
                      ) : null}
                    </div>
                  ) : (
                    <div className="text-muted-foreground flex h-full items-center justify-center">
                      <p className="text-sm">
                        Select a file from the tree to view details
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
