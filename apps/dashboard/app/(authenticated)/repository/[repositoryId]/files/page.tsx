"use client";

import { Files } from "lucide-react";
import { useState } from "react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  File,
  Folder,
  Tree,
  type TreeViewElement,
} from "@/components/ui/file-tree";
import { Button } from "@/components/ui/button";

const sampleFileTree: TreeViewElement[] = [
  {
    id: "1",
    name: "proto",
    children: [
      {
        id: "2",
        name: "user",
        children: [
          {
            id: "3",
            name: "user.proto",
          },
          {
            id: "4",
            name: "user_service.proto",
          },
        ],
      },
      {
        id: "5",
        name: "product",
        children: [
          {
            id: "6",
            name: "product.proto",
          },
          {
            id: "7",
            name: "product_service.proto",
          },
        ],
      },
      {
        id: "8",
        name: "common",
        children: [
          {
            id: "9",
            name: "types.proto",
          },
          {
            id: "10",
            name: "errors.proto",
          },
        ],
      },
    ],
  },
  {
    id: "11",
    name: "docs",
    children: [
      {
        id: "12",
        name: "README.md",
      },
      {
        id: "13",
        name: "API.md",
      },
    ],
  },
];

export default function FilesPage() {
  const [selectedFile, setSelectedFile] = useState<string | undefined>();

  const renderTree = (elements: TreeViewElement[]) => {
    return elements.map((element) => {
      if (element.children && element.children.length > 0) {
        return (
          <Folder key={element.id} element={element.name} value={element.id}>
            {renderTree(element.children)}
          </Folder>
        );
      }

      return (
        <File
          key={element.id}
          value={element.id}
          handleSelect={setSelectedFile}
        >
          <span>{element.name}</span>
        </File>
      );
    });
  };

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
                <div className="h-[500px] p-2">
                  <Tree
                    className="h-full w-full"
                    initialSelectedId={selectedFile}
                    elements={sampleFileTree}
                  >
                    {renderTree(sampleFileTree)}
                  </Tree>
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
                      <div className="rounded-lg border border-border bg-muted/30 p-4">
                        <p className="text-muted-foreground text-sm">
                          File preview will be displayed here for:{" "}
                          {selectedFile}
                        </p>
                      </div>
                      <div className="space-y-2">
                        <h4 className="text-sm font-medium">File Actions</h4>
                        <div className="flex flex-wrap gap-2">
                          <Button variant="outline" size="sm">
                            Download
                          </Button>
                          <Button variant="outline" size="sm">
                            Edit
                          </Button>
                          <Button variant="outline" size="sm">
                            History
                          </Button>
                          <Button variant="destructive" size="sm">
                            Delete
                          </Button>
                        </div>
                      </div>
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
