import { Visibility } from "@hasir/proto/gen/js/shared/visibility_pb";

type StrVisibility = "private" | "public";

export const visibilityMapper = new Map<StrVisibility, Visibility>([
    ["private", Visibility.PRIVATE],
    ["public", Visibility.PUBLIC],
]);

export const reverseVisibilityMapper = new Map<Visibility, StrVisibility>([
    [Visibility.PUBLIC, "public"],
    [Visibility.PRIVATE, "private"],
]);
