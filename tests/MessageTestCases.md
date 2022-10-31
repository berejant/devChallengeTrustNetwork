# Broadcast Message Tests (/api/message)

## Test1

1. Check that trust relation (edge) works.

   All persons should receive a Message. There are 8 persons.


    Sender -> Person1 -> Person4 -> Person6
    Sender -> Person2 -> Person5
    Sender -> Person3

    Person7, Person8 (which is not connected to a network),

Trust level is max. All have topics books.

Expected: 
 - message will be delivered to  Person1, Person2,  Person3, Person4 , Person5.
 - The message will not be delivered to Person7 or Person8.


## Test2
2. Check that inverse trust is not working

One topic (Cooking) for all People. Max trust level

    Sender -> Person1.
    Person2 -> Sender.

Expected: the message should be delivered only to Person1.

## Test3
3. Check trust connection levels

All people have topics swimming.

 - Sender -> ~~trust 3~~ to  Person1 -> trust 10 to Person4 -> trust 9 to Person6
 - Sender -> trust 10 to Person2 -> trust7 Person5
 - Sender -> trust 9 to Person3

Send a message with minimal trust level: 7

Expected:
  - message must be delivered only to Person2, Person5, Person3.
  - The message must not be delivered to  Person1, Person2, and Person5.

## Test4
4. All persons who receive a message must have appropriate topics.

Check, that person who has not the specific topics is ignored from broadcasting.
Trust level max for all persons.

  - Sender -> Person1 (cars, ~~sport~~) -> Person4 (cars, books, photography) -> Person6 (cars, books, photography)
  - Sender -> Person2 (cars, photography, swimming) -> Person5 (cars, photography)
  - Sender -> Person3 (cars, photography, football)

Send a message with topics: cars, photography

Expected:
  - message must be delivered only to Person2, Person5, Person3.
  - The message must not be delivered to Person1, Person4, and Person6

## Test5
5. Check trust connection levels + topics restriction
 - Sender -> ~~trust 3~~ to Person1 (cars, photography, football)  -> trust 10 to Person4 (cars, photography, football) -> trust 9 to Person6 (cars, sport)
 - Sender -> trust 10 to Person2 (cars, photography, football) -> trust7 Person5 (cars, photography)
 - Sender -> trust 9 to Person3 (cars, photography, painting)
 - Sender -> trust 8 Person6 (cars, ~~sport~~) -> trust 10 to Person7  (cars, photography)

Send a message with:
 * minimal trust level: 7
 * topics: cars, photography

Expected:
 * the message must be delivered only to Person2, Person5, and Person3.
 * the message must not be delivered only to: \
  Person1, Person2, Person5; \
  Person6, Person7, Person8

## Test6
6. Check Each person should receive this message only one time.

One topic (Cooking) for all People. Max trust level.
 - Sender -> Person1 - > Person2 -> Person 3
 - Sender -> Person2 -> Person3

Check that Person3 is present in response only one time.


# Not-broadcast Message  (/api/path)
## Test1

1. Check that the trust relation (edge) works. \
   Trust level is max.

 - Sender -> Person1 (no topics)  -> Person4 (books)
 - Sender -> Person2 ( painting )
 - Sender -> Person3 (books)

Find the shortest receiver with the topics book.
Expected: Person3, empty path.


## Test2
2. Check that inverse trust is not working

One topic (Cooking) for all Person3 and Person2. Max trust level
 - Sender -> Person1 -> Person3 (Cooking)
 - Person2 (Cooking) -> Sender.

Expected: the message should be delivered only to Person3 via (path) Person1.

## Test3
3. Check that the trust relation levels work.

Trust level is max.
 - Sender -> trust 7 to Person1 (no topics)  -> trust 8 to Person3 (swimming)
 - Sender -> ~~trust 3~~ to  Person2 (swimming)

Find the shortest receiver with the topics swimming.
Expected: Person3 with via Person1


## Test4
4. Check the shortest path
 - Sender -> Person1 -> Person4 -> Person5 (sport)
 - Sender -> Person2 -> Person5 (sport)
 - Sender -> Person3 -> Person6 -> Person4 -> Person5 (sport)

Find the shortest receiver with the topics sport.
Expected: Person5 with via Person2

## Test5
5. Check the shortest path with different trust level

 - Sender -> Person1 -> Person4 -> Person5 (sport)
 - Sender -> Person2 -> trust 5 to Person5 (sport)
 - Sender -> Person3 -> Person6 -> Person4 -> Person5 (sport)

Find the shortest receiver with the topics sport.
Expected: Person5 with via Person1, Person4

## Test6
6. Check there is no trust for matched people
 - Sender -> Person1;
 - Person2 -> trust 8 to Person3 (sport)

Find the shortest receiver with the topics sport.
Expected: no person
