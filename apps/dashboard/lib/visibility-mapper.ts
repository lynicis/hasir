import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";

export const visibilityMapper = new Map<"private" | "public", Visibility>([
    ["private", Visibility.PRIVATE],
    ["public", Visibility.PUBLIC],
]);
