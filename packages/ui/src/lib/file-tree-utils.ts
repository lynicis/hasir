export type TreeViewElement = {
  id: string;
  name: string;
  type?: "file" | "folder";
  isSelectable?: boolean;
  children?: TreeViewElement[];
  isLoaded?: boolean;
  isLoading?: boolean;
};

export type TreeSortMode =
  | "default"
  | "none"
  | ((a: TreeViewElement, b: TreeViewElement) => number);

export const isFolderElement = (element: TreeViewElement): boolean => {
  if (element.type) {
    return element.type === "folder";
  }

  return Array.isArray(element.children);
};

export const mergeExpandedItems = (
  currentItems: string[] | undefined,
  nextItems: string[]
): string[] => [...new Set([...(currentItems ?? []), ...nextItems])];

const treeCollator = new Intl.Collator("en", {
  numeric: true,
  sensitivity: "base",
});

export const defaultTreeComparator = (
  a: TreeViewElement,
  b: TreeViewElement
): number => {
  const aIsFolder = isFolderElement(a);
  const bIsFolder = isFolderElement(b);

  if (aIsFolder !== bIsFolder) {
    return aIsFolder ? -1 : 1;
  }

  return treeCollator.compare(a.name, b.name);
};

export const getTreeComparator = (sort: TreeSortMode) => {
  if (sort === "none") {
    return undefined;
  }

  if (sort === "default") {
    return defaultTreeComparator;
  }

  return sort;
};

export const sortTreeElements = (
  elements: TreeViewElement[],
  sort: TreeSortMode
): TreeViewElement[] => {
  const comparator = getTreeComparator(sort);

  const nextElements = elements.map((element) => {
    if (!Array.isArray(element.children)) {
      return element;
    }

    return {
      ...element,
      children: sortTreeElements(element.children, sort),
    };
  });

  if (!comparator) {
    return nextElements;
  }

  return [...nextElements].sort(comparator);
};
