import { ChevronLeftIcon, ChevronRightIcon, MoreHorizontalIcon } from "lucide-react"
import { Button } from "@hasir/ui/components/button"
import { cn } from "@hasir/ui/lib/utils"
import * as React from "react"

function PaginationRoot({ className, ...props }: React.ComponentProps<"nav">) {
  return (
    <nav
      role="navigation"
      aria-label="pagination"
      data-slot="pagination"
      className={cn("mx-auto flex w-full justify-center", className)}
      {...props}
    />
  )
}

function PaginationContent({
  className,
  ...props
}: React.ComponentProps<"ul">) {
  return (
    <ul
      data-slot="pagination-content"
      className={cn("flex items-center gap-0.5", className)}
      {...props}
    />
  )
}

function PaginationItem({ ...props }: React.ComponentProps<"li">) {
  return <li data-slot="pagination-item" {...props} />
}

type PaginationLinkProps = {
  isActive?: boolean
} & Omit<React.ComponentProps<typeof Button>, "variant">

function PaginationLink({
  className,
  isActive,
  size = "icon",
  ...props
}: PaginationLinkProps) {
  return (
    <Button
      variant={isActive ? "default" : "outline"}
      size={size}
      className={cn(className)}
      aria-current={isActive ? "page" : undefined}
      data-slot="pagination-link"
      data-active={isActive}
      {...props}
    />
  )
}

function PaginationPrevious({
  className,
  ...props
}: React.ComponentProps<typeof PaginationLink>) {
  return (
    <PaginationLink
      aria-label="Go to previous page"
      size="icon"
      className={cn("size-8 p-0", className)}
      {...props}
    >
      <ChevronLeftIcon />
    </PaginationLink>
  )
}

function PaginationNext({
  className,
  ...props
}: React.ComponentProps<typeof PaginationLink>) {
  return (
    <PaginationLink
      aria-label="Go to next page"
      size="icon"
      className={cn("size-8 p-0", className)}
      {...props}
    >
      <ChevronRightIcon />
    </PaginationLink>
  )
}

function PaginationEllipsis({
  className,
  ...props
}: React.ComponentProps<"span">) {
  return (
    <span
      aria-hidden
      data-slot="pagination-ellipsis"
      className={cn(
        "flex size-8 items-center justify-center [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    >
      <MoreHorizontalIcon
      />
      <span className="sr-only">More pages</span>
    </span>
  )
}

export interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  disabled?: boolean
  className?: string
}

const PAGE_DELTA = 2

type PageItem = number | "ellipsis"

export function getPageItems(
  currentPage: number,
  totalPages: number,
  delta = PAGE_DELTA
): PageItem[] {
  const windowSize = delta * 2 + 5
  if (!totalPages || totalPages <= 1) return []
  if (totalPages <= windowSize) {
    return Array.from({ length: totalPages }, (_, index) => index + 1)
  }

  const page = Math.min(Math.max(currentPage, 1), totalPages)
  const start = Math.max(2, page - delta)
  const end = Math.min(totalPages - 1, page + delta)
  const items: PageItem[] = [1]

  if (start > 2) items.push("ellipsis")
  for (let index = start; index <= end; index += 1) items.push(index)
  if (end < totalPages - 1) items.push("ellipsis")

  items.push(totalPages)
  return items
}

function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  disabled,
  className,
}: PaginationProps) {
  const pageItems = getPageItems(currentPage, totalPages)

  if (pageItems.length === 0) return null

  const page = Math.min(Math.max(currentPage, 1), totalPages)
  const isFirstPage = page === 1
  const isLastPage = page === totalPages

  return (
    <PaginationRoot className={className}>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={disabled || isFirstPage}
            className={disabled || isFirstPage ? "" : "cursor-pointer"}
          />
        </PaginationItem>
        {pageItems.map((item, index) =>
          item === "ellipsis" ? (
            <PaginationItem key={`ellipsis-${index}`}>
              <PaginationEllipsis />
            </PaginationItem>
          ) : (
            <PaginationItem key={item}>
              <PaginationLink
                onClick={() => onPageChange(item)}
                isActive={page === item}
                disabled={disabled}
                className={disabled ? "" : "cursor-pointer"}
              >
                {item.toString()}
              </PaginationLink>
            </PaginationItem>
          )
        )}
        <PaginationItem>
          <PaginationNext
            onClick={() => onPageChange(Math.min(totalPages, page + 1))}
            disabled={disabled || isLastPage}
            className={disabled || isLastPage ? "" : "cursor-pointer"}
          />
        </PaginationItem>
      </PaginationContent>
    </PaginationRoot>
  )
}

export {
  Pagination,
  PaginationRoot,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
}
