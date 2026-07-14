import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";

type StrVisibility = "private" | "public";

export const visibilityMapper = new Map<StrVisibility, Visibility>([
    ["private", Visibility.PRIVATE],
    ["public", Visibility.PUBLIC],
]);

export const reverseVisibilityMapper = new Map<Visibility, StrVisibility>([
    [Visibility.PUBLIC, "public"],
    [Visibility.PRIVATE, "private"],
]);
