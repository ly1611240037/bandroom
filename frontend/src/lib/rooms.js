export function filterRooms(rooms, people, keyword) {
  const query = keyword.trim().toLocaleLowerCase();
  return rooms.filter((room) => room.capacity >= Number(people || 0) && (
    !query || [room.name, ...(room.fixedEquipment || []).filter((item) => item.status === "available" && item.quantity > 0).map((item) => item.name)]
      .some((value) => value.toLocaleLowerCase().includes(query))
  ));
}
