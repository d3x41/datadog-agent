#ifndef _STRUCTS_DNS_H_
#define _STRUCTS_DNS_H_

struct dnshdr {
    uint16_t id;
    union {
        struct {
            uint8_t rd : 1;
            uint8_t tc : 1;
            uint8_t aa : 1;
            uint8_t opcode : 4;
            uint8_t qr : 1;

            uint8_t rcode : 4;
            uint8_t cd : 1;
            uint8_t ad : 1;
            uint8_t z : 1;
            uint8_t ra : 1;
        } as_bits_and_pieces;
        uint16_t as_value;
    } flags;
    uint16_t qdcount;
    uint16_t ancount;
    uint16_t nscount;
    uint16_t arcount;
};

struct dns_receiver_stats_t {
    // Packets that were filtered on the kernel
    u32 filtered_dns_packets;
    // Packets with the same ID and different size that didn't get filtered
    u32 same_id_different_size;
};

struct dns_responses_sent_to_userspace_lru_entry_t {
    u64 timestamp;
    u64 packet_size;
};

#endif
